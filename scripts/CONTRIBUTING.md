# Contributing to Real-time Chat System

Thank you for your interest in contributing! This document provides guidelines for contributing to the project.

## Getting Started

1. Fork the repository
2. Clone your fork
3. Create a feature branch
4. Make your changes
5. Submit a pull request

## Development Setup

See [QUICKSTART.md](QUICKSTART.md) for setup instructions.

## Code Style

### Backend (Go)

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting
- Run `go vet` before committing
- Add comments for exported functions
- Keep functions small and focused

### Frontend (TypeScript/React)

- Follow [TypeScript best practices](https://www.typescriptlang.org/docs/handbook/declaration-files/do-s-and-don-ts.html)
- Use functional components with hooks
- Keep components small and reusable
- Add TypeScript types for all props
- Use meaningful variable names

## Commit Messages

Follow [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: add user profile page
fix: resolve WebSocket reconnection issue
docs: update API documentation
style: format code with prettier
refactor: simplify message handling logic
test: add tests for auth service
chore: update dependencies
```

## Pull Request Process

1. **Update documentation** if needed
2. **Add tests** for new features
3. **Ensure all tests pass**
4. **Update CHANGELOG.md**
5. **Request review** from maintainers

### PR Title Format

```
[Type] Brief description

Example:
[Feature] Add message reactions
[Fix] Resolve memory leak in WebSocket handler
[Docs] Update deployment guide
```

### PR Description Template

```markdown
## Description
Brief description of changes

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
- [ ] Unit tests added/updated
- [ ] Integration tests added/updated
- [ ] Manual testing completed

## Checklist
- [ ] Code follows project style guidelines
- [ ] Self-review completed
- [ ] Comments added for complex code
- [ ] Documentation updated
- [ ] No new warnings generated
- [ ] Tests pass locally
```

## Testing Guidelines

### Backend Tests

```bash
# Run all tests
make test

# Run specific test
go test ./internal/service -v

# Run with coverage
go test -cover ./...
```

### Frontend Tests

```bash
cd frontend

# Build test
npm run build

# Lint
npm run lint
```

## Feature Requests

1. Check existing issues first
2. Create a new issue with:
   - Clear description
   - Use cases
   - Expected behavior
   - Mockups (if UI change)

## Bug Reports

Include:
- Steps to reproduce
- Expected behavior
- Actual behavior
- Screenshots (if applicable)
- Environment details
- Error messages/logs

## Code Review Process

1. Maintainers review PRs
2. Address feedback
3. Get approval from at least one maintainer
4. Merge when approved

## Areas for Contribution

### High Priority
- [ ] Redis integration for scaling
- [ ] Message search functionality
- [ ] File upload support
- [ ] User blocking/reporting
- [ ] Admin dashboard

### Medium Priority
- [ ] Message reactions
- [ ] User profiles
- [ ] Settings page
- [ ] Dark mode
- [ ] Mobile app

### Low Priority
- [ ] Voice/video calls
- [ ] Message encryption
- [ ] Analytics dashboard
- [ ] Bot integration

## Documentation

- Update README.md for user-facing changes
- Update docs/ for technical changes
- Add inline comments for complex logic
- Update API documentation

## Questions?

- Open an issue for questions
- Join our community chat (if available)
- Email maintainers

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

## Code of Conduct

- Be respectful and inclusive
- Welcome newcomers
- Focus on constructive feedback
- Help others learn

## Recognition

Contributors will be:
- Listed in CONTRIBUTORS.md
- Mentioned in release notes
- Credited in documentation

Thank you for contributing! 🎉
