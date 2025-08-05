You are building the project detailed in README.md. Follow these implementation instructions:

- Use Go for implementation.
    - Create a cli entry point in cli/.
    - Create separate entry points in other locations.
- Prioritize creating an intuitive CLI experience.
    - Give good error messages.
    - Create flags with a `--` prefix and a shorthand with a `-` prefix.
- Create regression tests for every feature as they are implemented. 
    - Store tests in test/.
- Use git rigorously.
    - Use feature branches as you work to keep the project organized.
    - Typecheck must pass before creating a commit.
    - All tests must pass before creating a commit.
    - When adding files to a commit, specify EVERY file you intend to add.
    - Commit messages MUST be 1 sentence long.
- Use a Makefile for building, cleaning, and testing.
    - Create scripts as you go.
    - For example, create and use `make clean`, `make test`, and `make all`.
- Take notes in CLAUDE.md about implementation details.
    - For example, take notes on:
        - The data structures you used.
        - The libraries you used.
        - The assumptions you made.
        - The features you implemented and the ones that you pushed off for later.

NOTES:

