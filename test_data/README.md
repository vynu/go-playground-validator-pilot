# Test Data Directory

This directory contains test data files used by the e2e_test_suite.sh for comprehensive testing of the Go Playground Data Validator.

## Directory Structure

```
test_data/
├── valid/          # Valid payloads that should pass validation
├── invalid/        # Invalid payloads that should fail validation
├── examples/       # Example payloads for reference and extended testing
└── README.md       # This file
```

## File Naming Convention

Test data files should follow the naming pattern: `{model_name}.json`

Examples:
- `valid/incident.json` - Valid incident payload
- `invalid/incident.json` - Invalid incident payload
- `examples/github.json` - Example GitHub webhook payload

## Adding New Test Data

### For Existing Models

1. **Valid Payload**: Create `valid/{model_name}.json` with a payload that should pass validation
2. **Invalid Payload**: Create `invalid/{model_name}.json` with a payload that should fail validation
3. **Examples**: Optionally create `examples/{model_name}.json` for reference

### For New Models

When adding a new model to the system:

1. Create the model files (`src/models/{model_name}.go` and `src/validations/{model_name}.go`)
2. Add test data files:
   - `test_data/valid/{model_name}.json`
   - `test_data/invalid/{model_name}.json`
3. The e2e_test_suite.sh will automatically discover and test the new model

## File Format

All test data files must be valid JSON. Example structure:

```json
{
  "field1": "value1",
  "field2": 123,
  "field3": true,
  "field4": ["array", "values"],
  "field5": {
    "nested": "object"
  }
}
```

## Validation Guidelines

### Valid Payloads
- Should contain all required fields
- Should use correct data types
- Should satisfy all validation rules (min/max lengths, formats, etc.)
- Should represent realistic use cases

### Invalid Payloads
- Should violate one or more validation rules
- Common invalid scenarios:
  - Missing required fields
  - Wrong data types
  - Values outside allowed ranges
  - Invalid formats (email, dates, etc.)
  - Empty strings where content is required

## Testing Specific Scenarios

### Edge Cases
Create additional test files with descriptive suffixes:
- `valid/{model_name}_edge_case.json`
- `invalid/{model_name}_boundary.json`

### Performance Testing
For large payload testing:
- `examples/{model_name}_large.json`

### Unicode/International
For internationalization testing:
- `valid/{model_name}_unicode.json`

## Usage by e2e_test_suite.sh

The test suite automatically:
1. Discovers all models in the system
2. Looks for corresponding test data files in `test_data/valid/` and `test_data/invalid/`
3. Tests validation with found payloads
4. Reports results for each model

Manual test data files in this directory take precedence over hardcoded payloads in the test script.