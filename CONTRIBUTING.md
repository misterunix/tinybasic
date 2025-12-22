# Contributing to TinyBASIC

Thank you for your interest in contributing to TinyBASIC! This document provides guidelines for contributing to the project.

## How to Contribute

### Reporting Bugs

If you find a bug, please open an issue on GitHub with:
- A clear, descriptive title
- Steps to reproduce the issue
- Expected behavior vs actual behavior
- BASIC code sample that demonstrates the problem
- Your Go version and operating system

### Suggesting Enhancements

Enhancement suggestions are welcome! Please open an issue with:
- A clear description of the enhancement
- Why this enhancement would be useful
- Example use cases
- If possible, example BASIC code showing the desired behavior

### Pull Requests

1. **Fork the repository** and create your branch from `main`
2. **Make your changes** following the code style guidelines below
3. **Add tests** if applicable
4. **Update documentation** if you're changing functionality
5. **Ensure all tests pass** by running `go test ./...`
6. **Submit a pull request** with a clear description of your changes

#### Pull Request Guidelines

- Keep pull requests focused on a single feature or fix
- Write clear, descriptive commit messages
- Reference any related issues in your PR description
- Be responsive to feedback and be prepared to make changes

## Code Style Guidelines

- Follow standard Go formatting using `gofmt`
- Use meaningful variable and function names
- Add comments for complex logic
- Keep functions focused and reasonably sized
- Maintain the existing code structure and patterns

## Development Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/misterunix/tinybasic.git
   cd tinybasic
   ```

2. Ensure you have Go 1.20 or later installed:
   ```bash
   go version
   ```

3. Run tests:
   ```bash
   go test ./...
   ```

## Testing

- Write tests for new features
- Ensure existing tests pass
- Include both positive and negative test cases
- Test edge cases and error conditions

## Documentation

- Update README.md if adding new features
- Add inline comments for complex code
- Include usage examples for new functionality
- Keep the API documentation up to date

## BASIC Language Features

When adding new BASIC statements or functions:
- Follow traditional BASIC conventions where appropriate
- Ensure compatibility with existing features
- Update the feature list in documentation
- Provide example code demonstrating the feature

## License

By contributing to TinyBASIC, you agree that your contributions will be licensed under the BSD 3-Clause License.

## Questions?

Feel free to open an issue for any questions about contributing!

## Code of Conduct

Please note that this project is released with a [Contributor Code of Conduct](CODE_OF_CONDUCT.md). By participating in this project you agree to abide by its terms.
