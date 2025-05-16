# Contributing to Student Code Viewer (scv)

First off, thank you for considering contributing to scv! It's people like you that make scv such a great tool.

## How Can I Contribute?

There are many ways you can contribute to the project:

*   **Reporting Bugs**: If you find a bug, please open an issue and describe it in detail.
*   **Suggesting Enhancements**: If you have an idea for a new feature or an improvement to an existing one, open an issue to discuss it.
*   **Writing Code**: If you're up for writing some Go, feel free to pick up an open issue or propose a new feature.
*   **Improving Documentation**: Clear documentation is vital. If you see areas for improvement, let us know or submit a pull request.

## Getting Started

To contribute code, you'll need to set up your development environment:

1.  **Install Go**: Ensure you have Go installed (version 1.20 or newer is recommended). You can download it from [golang.org](https://golang.org/dl/).
2.  **Clone the Repository**:
    ```bash
    git clone https://github.com/asp2131/student-code-viewer.git
    cd student-code-viewer
    ```
3.  **Install Dependencies**:
    This project uses Go modules, so dependencies should be automatically managed. You can ensure they are downloaded by running:
    ```bash
    go mod tidy
    ```

## Building the Project

To build the `scv` executable from source:

```bash
go build -o scv .
```
This will create an `scv` executable in the project's root directory.

## Running Tests

(To be added: Instructions on how to run project tests, if any are set up.)

## Coding Standards

*   Follow standard Go formatting (`gofmt`).
*   Write clear and concise commit messages.
*   Comment your code where necessary, especially for complex logic.

## Pull Request Process

1.  Ensure any install or build dependencies are removed before the end of the layer when doing a build.
2.  Update the README.md with details of changes to the interface, this includes new environment variables, exposed ports, useful file locations and container parameters.
3.  Increase the version numbers in any examples and the README.md to the new version that this Pull Request would represent. The versioning scheme we use is [SemVer](http://semver.org/).
4.  You may merge the Pull Request in once you have the sign-off of two other developers, or if you do not have permission to do that, you may request the second reviewer to merge it for you.

## Code of Conduct

Please note that this project is released with a Contributor Code of Conduct. By participating in this project you agree to abide by its terms. (We can add a `CODE_OF_CONDUCT.md` file later if you'd like).

We look forward to your contributions!
