[![License: GPL v3](https://img.shields.io/badge/License-GPL%20v3-blue.svg)](http://www.gnu.org/licenses/gpl-3.0)
[![Documentation](https://img.shields.io/badge/Documentation-GoDoc-green.svg)](https://godoc.org/github.com/dylhunn/dragontooth)

Dragontooth | Dylan D. Hunn
===========================

Dragontooth is a fast, UCI-compliant chess engine written in Go.

Repo summary
============

Here is a summary of the important files in the repo:

| **File**         | **Description**                                                                                                                                         |
|--------------|------------------------------------------------------------------------------------------------------------------------------------------------------|
| main.go       | This is the UCI entrypoint that handles communication with the GUI. |

**This project is currently a concept.** Check back soon for progress.

Installing and building Dragontooth
===================================

This project requires Go 1.9. As of the time of writing, 1.9 is still a pre-release version. You can get it by cloning the official [Go Repo](https://github.com/golang/go), and building it yourself.

To build Dragontooth from source, make sure your `GO_PATH` environment variable is correctly set, and install it using `go get`:

    go get github.com/dylhunn/dragontooth

Alternatively, you can clone it yourself, but this will require you to clone [the dependency](https://github.com/dylhunn/dragontoothmg) as well, and configure them at the correct paths:

    git clone https://github.com/dylhunn/dragontooth.git
    git clone https://github.com/dylhunn/dragontoothmg.git

Documentation
=============

You can find the documentation [here](https://godoc.org/github.com/dylhunn/dragontooth).