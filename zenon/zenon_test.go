package zenon

import (
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/0x3639/znn-sdk-go/api/embedded"
	"github.com/0x3639/znn-sdk-go/pow"
	"github.com/0x3639/znn-sdk-go/rpc_client"
	"github.com/0x3639/znn-sdk-go/transport"
	"github.com/0x3639/znn-sdk-go/utils"
	"github.com/0x3639/znn-sdk-go/wallet"
	"github.com/zenon-network/go-zenon/chain/nom"
	"github.com/zenon-network/go-zenon/common/types"
	gozenonpow "github.com/zenon-network/go-zenon/pow"
	nodeapi "github.com/zenon-network/go-zenon/rpc/api"
)

// testMnemonic is a well-known valid BIP39 mnemonic used only for deterministic
// offline tests.
const testMnemonic = "test test test test test test test test test test test junk"

func testKeyPair(t *testing.T) *wallet.KeyPair {
	t.Helper()
	ks, err := wallet.NewKeyStoreFromMnemonic(testMnemonic)
	if err != nil {
		t.Fatalf("NewKeyStoreFromMnemonic: %v", err)
	}
	kp, err := ks.GetKeyPair(0)
	if err != nil {
		t.Fatalf("GetKeyPair: %v", err)
	}
	return kp
}

func sampleSendBlock(t *testing.T, kp *wallet.KeyPair) *nom.AccountBlock {
	t.Helper()
	addr, err := kp.GetAddress()
	if err != nil {
		t.Fatalf("GetAddress: %v", err)
	}
	pub, err := kp.GetPublicKey()
	if err != nil {
		t.Fatalf("GetPublicKey: %v", err)
	}
	return &nom.AccountBlock{
		Version:              1,
		ChainIdentifier:      1,
		BlockType:            nom.BlockTypeUserSend,
		Address:              *addr,
		ToAddress:            types.ParseAddressPanic("z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz"),
		Amount:               big.NewInt(100000000),
		TokenStandard:        types.ZnnTokenStandard,
		Height:               5,
		PreviousHash:         types.HexToHashPanic("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"),
		MomentumAcknowledged: types.HashHeight{Hash: types.ZeroHash, Height: 100},
		PublicKey:            pub,
	}
}

// TestSetHashAndSignature verifies that the signing step produces the canonical
// transaction hash and a signature the keypair can verify.
func TestSetHashAndSignature(t *testing.T) {
	kp := testKeyPair(t)
	block := sampleSendBlock(t, kp)

	z := &Zenon{} // no client needed for signing
	if err := z.setHashAndSignature(block, kp); err != nil {
		t.Fatalf("setHashAndSignature: %v", err)
	}

	wantHash := utils.GetTransactionHash(block)
	if block.Hash != wantHash {
		t.Errorf("block.Hash = %s, want %s", block.Hash, wantHash)
	}

	ok, err := kp.Verify(block.Signature, block.Hash.Bytes())
	if err != nil {
		t.Fatalf("Verify: %v", err)
	}
	if !ok {
		t.Error("signature does not verify against the transaction hash")
	}
}

// TestNormalizeBlockDefaultsSend verifies that a send template gets a default
// protocol version and a non-nil amount while its routing fields are preserved.
func TestNormalizeBlockDefaultsSend(t *testing.T) {
	to := types.ParseAddressPanic("z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz")
	block := &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserSend,
		ToAddress:     to,
		TokenStandard: types.ZnnTokenStandard,
		Amount:        big.NewInt(42),
		// Version intentionally left at 0 to exercise the default.
	}

	normalizeBlockDefaults(block)

	if block.Version != 1 {
		t.Errorf("Version = %d, want 1", block.Version)
	}
	if block.Amount == nil || block.Amount.Cmp(big.NewInt(42)) != 0 {
		t.Errorf("Amount = %v, want 42", block.Amount)
	}
	if block.ToAddress != to {
		t.Errorf("ToAddress = %s, want %s (send routing must be preserved)", block.ToAddress, to)
	}
	if block.TokenStandard != types.ZnnTokenStandard {
		t.Errorf("TokenStandard = %s, want ZNN (send routing must be preserved)", block.TokenStandard)
	}
}

// TestNormalizeBlockDefaultsReceive verifies that a receive template gets a
// non-nil zero amount and that any stray routing fields are zeroed, matching the
// node's receive-block verification (ErrABAmountMustBeZero/ZtsMustBeZero/
// ToAddressMustBeZero) and the TypeScript SDK defaults.
func TestNormalizeBlockDefaultsReceive(t *testing.T) {
	block := &nom.AccountBlock{
		BlockType:     nom.BlockTypeUserReceive,
		FromBlockHash: types.HexToHashPanic("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"),
		// Stray routing fields that a dirty template might carry.
		ToAddress:     types.ParseAddressPanic("z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz"),
		TokenStandard: types.ZnnTokenStandard,
		// Amount left nil to exercise normalization.
	}

	normalizeBlockDefaults(block)

	if block.Version != 1 {
		t.Errorf("Version = %d, want 1", block.Version)
	}
	if block.Amount == nil || block.Amount.Sign() != 0 {
		t.Errorf("Amount = %v, want non-nil zero", block.Amount)
	}
	if block.ToAddress != types.ZeroAddress {
		t.Errorf("ToAddress = %s, want zero for receive block", block.ToAddress)
	}
	if block.TokenStandard != types.ZeroTokenStandard {
		t.Errorf("TokenStandard = %s, want zero for receive block", block.TokenStandard)
	}
}

// TestNormalizeBlockDefaultsPreservesVersion confirms a caller-supplied version
// is not overwritten.
func TestNormalizeBlockDefaultsPreservesVersion(t *testing.T) {
	block := &nom.AccountBlock{BlockType: nom.BlockTypeUserSend, Version: 2}
	normalizeBlockDefaults(block)
	if block.Version != 2 {
		t.Errorf("Version = %d, want 2 (caller value must be preserved)", block.Version)
	}
}

// TestSendFlowNonceAcceptedByNode confirms the nonce that the send flow would
// generate for a block satisfies go-zenon's pow.CheckPoWNonce. This guards the
// integration between setDifficulty's data hash and the pow package.
func TestSendFlowNonceAcceptedByNode(t *testing.T) {
	kp := testKeyPair(t)
	block := sampleSendBlock(t, kp)
	block.Difficulty = 1000

	dataHash := gozenonpow.GetAccountBlockHash(block)
	nonce := pow.GeneratePowBytes(dataHash, block.Difficulty)
	copy(block.Nonce.Data[:], nonce)

	if !gozenonpow.CheckPoWNonce(block) {
		t.Errorf("send-flow nonce %x rejected by go-zenon CheckPoWNonce", nonce)
	}
}

type zenonRPCFixture struct {
	frontier  interface{}
	momentum  interface{}
	source    interface{}
	pow       embedded.GetRequiredResult
	errors    map[string]string
	calls     []string
	published *nom.AccountBlock
}

func newZenonTestClient(t *testing.T, fixture *zenonRPCFixture) (*rpc_client.RpcClient, func()) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer request.Body.Close()
		var rpcRequest transport.Request
		if err := json.NewDecoder(request.Body).Decode(&rpcRequest); err != nil {
			t.Errorf("decode request: %v", err)
			writer.WriteHeader(http.StatusBadRequest)
			return
		}
		fixture.calls = append(fixture.calls, rpcRequest.Method)
		writer.Header().Set("Content-Type", "application/json")
		if message := fixture.errors[rpcRequest.Method]; message != "" {
			_ = json.NewEncoder(writer).Encode(map[string]interface{}{
				"jsonrpc": "2.0", "id": rpcRequest.ID,
				"error": map[string]interface{}{"code": -32000, "message": message},
			})
			return
		}
		var result interface{}
		switch rpcRequest.Method {
		case "ledger.getFrontierAccountBlock":
			result = fixture.frontier
		case "ledger.getFrontierMomentum":
			result = fixture.momentum
		case "ledger.getAccountBlockByHash":
			result = fixture.source
		case "embedded.plasma.getRequiredPoWForAccountBlock":
			result = fixture.pow
		case "ledger.publishRawTransaction":
			if len(rpcRequest.Params) == 1 {
				raw, _ := json.Marshal(rpcRequest.Params[0])
				fixture.published = new(nom.AccountBlock)
				_ = json.Unmarshal(raw, fixture.published)
			}
			result = nil
		default:
			t.Errorf("unexpected RPC method %q", rpcRequest.Method)
		}
		_ = json.NewEncoder(writer).Encode(map[string]interface{}{
			"jsonrpc": "2.0", "id": rpcRequest.ID, "result": result,
		})
	}))

	options := rpc_client.DefaultClientOptions()
	options.AutoReconnect = false
	options.HealthCheckInterval = 0
	client, err := rpc_client.NewRpcClientWithOptions(server.URL, options)
	if err != nil {
		server.Close()
		t.Fatalf("NewRpcClientWithOptions: %v", err)
	}
	cleanup := func() {
		client.Stop()
		server.Close()
	}
	return client, cleanup
}

func testMomentum(height, chainIdentifier uint64, hash types.Hash) *nodeapi.Momentum {
	return &nodeapi.Momentum{Momentum: &nom.Momentum{
		Version:         1,
		ChainIdentifier: chainIdentifier,
		Hash:            hash,
		Height:          height,
	}}
}

func TestZenonSendCompletesPlasmaBackedFlow(t *testing.T) {
	momentumHash := types.HexToHashPanic("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	fixture := &zenonRPCFixture{
		momentum: testMomentum(99, 7, momentumHash),
		pow:      embedded.GetRequiredResult{BasePlasma: 21000},
		errors:   make(map[string]string),
	}
	client, cleanup := newZenonTestClient(t, fixture)
	defer cleanup()

	z := NewZenon(client)
	if z.Client() != client {
		t.Fatal("Client() did not return the configured RPC client")
	}
	kp := testKeyPair(t)
	to := types.ParseAddressPanic("z1qzal6c5s9rjnnxd2z7dvdhjxpmmj4fmw56a0mz")
	template := client.LedgerApi.SendTemplate(to, types.ZnnTokenStandard, big.NewInt(42), []byte("memo"))

	published, err := z.Send(template, kp)
	if err != nil {
		t.Fatalf("Send: %v", err)
	}
	if published != template || fixture.published == nil {
		t.Fatal("Send did not publish the prepared template")
	}
	if template.Height != 1 || template.PreviousHash != types.ZeroHash || template.ChainIdentifier != 7 {
		t.Fatalf("chain position = height %d previous %s chain %d", template.Height, template.PreviousHash, template.ChainIdentifier)
	}
	if template.MomentumAcknowledged.Hash != momentumHash || template.MomentumAcknowledged.Height != 99 {
		t.Fatalf("momentum acknowledgment = %+v", template.MomentumAcknowledged)
	}
	if template.FusedPlasma != 21000 || template.Difficulty != 0 || template.Nonce.Data != ([8]byte{}) {
		t.Fatalf("plasma fields = fused %d difficulty %d nonce %x", template.FusedPlasma, template.Difficulty, template.Nonce.Data)
	}
	if len(template.PublicKey) == 0 || len(template.Signature) == 0 || template.Hash == types.ZeroHash {
		t.Fatal("prepared transaction is missing signing fields")
	}
	wantCalls := []string{
		"ledger.getFrontierAccountBlock",
		"ledger.getFrontierMomentum",
		"embedded.plasma.getRequiredPoWForAccountBlock",
		"ledger.publishRawTransaction",
	}
	if !reflect.DeepEqual(fixture.calls, wantCalls) {
		t.Fatalf("RPC calls = %v, want %v", fixture.calls, wantCalls)
	}
}

func TestZenonPrepareBlockGeneratesPoWAndPreservesChainID(t *testing.T) {
	frontierHash := types.HexToHashPanic("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	momentumHash := types.HexToHashPanic("cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc")
	fixture := &zenonRPCFixture{
		frontier: &nodeapi.AccountBlock{AccountBlock: nom.AccountBlock{Height: 7, Hash: frontierHash, Amount: big.NewInt(0)}},
		momentum: testMomentum(100, 9, momentumHash),
		pow:      embedded.GetRequiredResult{AvailablePlasma: 11, BasePlasma: 22, RequiredDifficulty: 1},
		errors:   make(map[string]string),
	}
	client, cleanup := newZenonTestClient(t, fixture)
	defer cleanup()
	z := NewZenon(client)
	var statuses []pow.PowStatus
	z.PowCallback = func(status pow.PowStatus) { statuses = append(statuses, status) }

	kp := testKeyPair(t)
	template := client.LedgerApi.SendTemplate(types.PlasmaContract, types.QsrTokenStandard, big.NewInt(1), nil)
	template.ChainIdentifier = 77
	prepared, err := z.PrepareBlock(template, kp)
	if err != nil {
		t.Fatalf("PrepareBlock: %v", err)
	}
	if prepared != template || template.Height != 8 || template.PreviousHash != frontierHash || template.ChainIdentifier != 77 {
		t.Fatalf("prepared chain position = %+v", template)
	}
	if template.FusedPlasma != 11 || template.Difficulty != 1 || !gozenonpow.CheckPoWNonce(template) {
		t.Fatalf("PoW fields = fused %d difficulty %d nonce %x", template.FusedPlasma, template.Difficulty, template.Nonce.Data)
	}
	if !reflect.DeepEqual(statuses, []pow.PowStatus{pow.Generating, pow.Done}) {
		t.Fatalf("PoW statuses = %v", statuses)
	}

	required, err := z.RequiresPoW(client.LedgerApi.SendTemplate(types.PlasmaContract, types.ZnnTokenStandard, big.NewInt(0), nil), kp)
	if err != nil || !required {
		t.Fatalf("RequiresPoW = %v, %v", required, err)
	}
}

func TestZenonFlowValidationAndRPCFailures(t *testing.T) {
	momentum := testMomentum(1, 1, types.ZeroHash)
	address, err := testKeyPair(t).GetAddress()
	if err != nil {
		t.Fatal(err)
	}
	validSource := &nodeapi.AccountBlock{AccountBlock: nom.AccountBlock{ToAddress: *address, Amount: big.NewInt(1)}}

	tests := []struct {
		name    string
		fixture *zenonRPCFixture
		block   func() *nom.AccountBlock
		want    string
	}{
		{
			name:    "frontier error",
			fixture: &zenonRPCFixture{momentum: momentum, errors: map[string]string{"ledger.getFrontierAccountBlock": "frontier failed"}},
			block:   func() *nom.AccountBlock { return &nom.AccountBlock{BlockType: nom.BlockTypeUserSend} },
			want:    "failed to get frontier account block",
		},
		{
			name:    "momentum error",
			fixture: &zenonRPCFixture{momentum: momentum, errors: map[string]string{"ledger.getFrontierMomentum": "momentum failed"}},
			block:   func() *nom.AccountBlock { return &nom.AccountBlock{BlockType: nom.BlockTypeUserSend} },
			want:    "failed to get frontier momentum",
		},
		{
			name:    "momentum unavailable",
			fixture: &zenonRPCFixture{momentum: nil, errors: make(map[string]string)},
			block:   func() *nom.AccountBlock { return &nom.AccountBlock{BlockType: nom.BlockTypeUserSend} },
			want:    "frontier momentum unavailable",
		},
		{
			name:    "receive missing source hash",
			fixture: &zenonRPCFixture{momentum: momentum, errors: make(map[string]string)},
			block:   func() *nom.AccountBlock { return &nom.AccountBlock{BlockType: nom.BlockTypeUserReceive} },
			want:    "non-empty fromBlockHash",
		},
		{
			name:    "receive source query error",
			fixture: &zenonRPCFixture{momentum: momentum, errors: map[string]string{"ledger.getAccountBlockByHash": "source failed"}},
			block: func() *nom.AccountBlock {
				return &nom.AccountBlock{BlockType: nom.BlockTypeUserReceive, FromBlockHash: types.HexToHashPanic("dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd")}
			},
			want: "failed to fetch source send block",
		},
		{
			name:    "receive wrong recipient",
			fixture: &zenonRPCFixture{momentum: momentum, source: &nodeapi.AccountBlock{AccountBlock: nom.AccountBlock{ToAddress: types.PlasmaContract, Amount: big.NewInt(1)}}, errors: make(map[string]string)},
			block: func() *nom.AccountBlock {
				return &nom.AccountBlock{BlockType: nom.BlockTypeUserReceive, FromBlockHash: types.HexToHashPanic("eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee")}
			},
			want: "does not match account",
		},
		{
			name:    "receive data",
			fixture: &zenonRPCFixture{momentum: momentum, source: validSource, errors: make(map[string]string)},
			block: func() *nom.AccountBlock {
				return &nom.AccountBlock{BlockType: nom.BlockTypeUserReceive, FromBlockHash: types.HexToHashPanic("ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"), Data: []byte("not allowed")}
			},
			want: "must not carry data",
		},
		{
			name:    "prefilled difficulty without nonce",
			fixture: &zenonRPCFixture{momentum: momentum, errors: make(map[string]string)},
			block:   func() *nom.AccountBlock { return &nom.AccountBlock{BlockType: nom.BlockTypeUserSend, Difficulty: 1} },
			want:    "but no nonce",
		},
		{
			name:    "pow query error",
			fixture: &zenonRPCFixture{momentum: momentum, errors: map[string]string{"embedded.plasma.getRequiredPoWForAccountBlock": "pow failed"}},
			block:   func() *nom.AccountBlock { return &nom.AccountBlock{BlockType: nom.BlockTypeUserSend} },
			want:    "failed to query required PoW",
		},
		{
			name:    "hostile difficulty",
			fixture: &zenonRPCFixture{momentum: momentum, pow: embedded.GetRequiredResult{RequiredDifficulty: pow.MaxReasonableDifficulty + 1}, errors: make(map[string]string)},
			block:   func() *nom.AccountBlock { return &nom.AccountBlock{BlockType: nom.BlockTypeUserSend} },
			want:    "above the maximum supported",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, cleanup := newZenonTestClient(t, test.fixture)
			defer cleanup()
			_, err := NewZenon(client).PrepareBlock(test.block(), testKeyPair(t))
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("error = %v, want substring %q", err, test.want)
			}
		})
	}
}

func TestZenonSendWrapsPublishFailure(t *testing.T) {
	fixture := &zenonRPCFixture{
		momentum: testMomentum(1, 1, types.ZeroHash),
		errors:   map[string]string{"ledger.publishRawTransaction": "publish failed"},
	}
	client, cleanup := newZenonTestClient(t, fixture)
	defer cleanup()
	block := client.LedgerApi.SendTemplate(types.PlasmaContract, types.ZnnTokenStandard, big.NewInt(1), nil)
	if _, err := NewZenon(client).Send(block, testKeyPair(t)); err == nil || !strings.Contains(err.Error(), "failed to publish transaction") {
		t.Fatalf("Send error = %v", err)
	}
}
