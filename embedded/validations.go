package embedded

import "fmt"

// =============================================================================
// Token Validations
// =============================================================================

// ValidateTokenName validates a token name according to protocol rules
func ValidateTokenName(value string) error {
	if value == "" {
		return fmt.Errorf("token name cannot be empty")
	}

	if !TokenNameRegExp.MatchString(value) {
		return fmt.Errorf("token name must contain only alphanumeric characters")
	}

	if len(value) > TokenNameMaxLength {
		return fmt.Errorf("token name must have maximum %d characters", TokenNameMaxLength)
	}

	return nil
}

// ValidateTokenSymbol validates a token symbol according to protocol rules
func ValidateTokenSymbol(value string) error {
	if value == "" {
		return fmt.Errorf("token symbol cannot be empty")
	}

	if !TokenSymbolRegExp.MatchString(value) {
		return fmt.Errorf("token symbol must match pattern: %s", TokenSymbolRegExp.String())
	}

	if len(value) > TokenSymbolMaxLength {
		return fmt.Errorf("token symbol must have maximum %d characters", TokenSymbolMaxLength)
	}

	// Check if symbol is reserved
	for _, reserved := range TokenSymbolExceptions {
		if value == reserved {
			return fmt.Errorf("token symbol must not be one of the following: %v", TokenSymbolExceptions)
		}
	}

	return nil
}

// ValidateTokenDomain validates a token domain according to protocol rules
func ValidateTokenDomain(value string) error {
	if value == "" {
		return fmt.Errorf("token domain cannot be empty")
	}

	if !TokenDomainRegExp.MatchString(value) {
		return fmt.Errorf("domain is not valid")
	}

	return nil
}

// =============================================================================
// Pillar Validations
// =============================================================================

// ValidatePillarName validates a pillar name according to protocol rules
func ValidatePillarName(value string) error {
	if value == "" {
		return fmt.Errorf("pillar name cannot be empty")
	}

	if !PillarNameRegExp.MatchString(value) {
		return fmt.Errorf("pillar name must match pattern: %s", PillarNameRegExp.String())
	}

	if len(value) > PillarNameMaxLength {
		return fmt.Errorf("pillar name must have maximum %d characters", PillarNameMaxLength)
	}

	return nil
}

// =============================================================================
// Accelerator Project Validations
// =============================================================================

// ValidateProjectName validates an accelerator project name
func ValidateProjectName(value string) error {
	if value == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	if len(value) > ProjectNameMaxLength {
		return fmt.Errorf("project name must have maximum %d characters", ProjectNameMaxLength)
	}

	return nil
}

// ValidateProjectDescription validates an accelerator project description
func ValidateProjectDescription(value string) error {
	if value == "" {
		return fmt.Errorf("project description cannot be empty")
	}

	if len(value) > ProjectDescriptionMaxLength {
		return fmt.Errorf("project description must have maximum %d characters", ProjectDescriptionMaxLength)
	}

	return nil
}
