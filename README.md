# Automountify

Automountify is a terminal-based utility written in Go that allows users to select a disk, format it with the `ext4` filesystem, create a mount point, assign permissions, and mount the disk. It also automatically updates the `/etc/fstab` file for persistent mounting.

This utility provides a simple, interactive interface using the `bubbletea` package for building TUI (Text User Interface) applications.

## Features

- **Disk Selection**: Displays available disks for formatting and mounting.
- **Mount Point Setup**: Allows the user to input a custom mount point (e.g., `/mnt/data`).
- **Permissions Setup**: Lets the user specify permissions (e.g., `0755`) for the mount point directory.
- **Formatting**: Formats the selected disk with the `ext4` filesystem.
- **Mounting**: Mounts the disk to the specified mount point.
- **Persistent Mounting**: Adds the new mount entry to the `/etc/fstab` file to ensure it persists across reboots.
- **Loading Indicator**: Displays a loading spinner during disk formatting and mounting.
- **Error Handling**: Provides detailed error messages if any step fails.

## Prerequisites

Before using this utility, ensure that you have the following installed:

- **Go** (v1.18 or higher) - Go programming language environment
- **sudo** privileges - Required for running commands like `mkfs.ext4`, `mount`, and updating `/etc/fstab`
- **bubbletea** - A Go package for creating TUI applications
  - Install it with the following:
    ```bash
    go get github.com/charmbracelet/bubbletea
    ```

## Installation

To install and run this utility:

1. Clone this repository:
    ```bash
    git clone https://github.com/kintsdev/automountify.git
    cd automountify
    ```

2. Install the necessary Go dependencies:
    ```bash
    go mod tidy
    ```

3. Build the application:
    ```bash
    go build -o automountify
    ```

4. Run the application:
    ```bash
    sudo ./automountify
    ```

> **Note:** Running the program requires `sudo` privileges, as it performs system-level operations like formatting a disk and modifying the `/etc/fstab` file.

## Usage

1. **Select a Disk**: The program will list available disks. Use the arrow keys to select a disk and press Enter to proceed.
2. **Enter Mount Point**: After selecting a disk, enter the mount point where the disk should be mounted (e.g., `/mnt/data`).
3. **Enter Permissions**: Enter the permissions for the mount point directory (e.g., `0755`).
4. **Format and Mount**: The program will format the disk with the `ext4` filesystem, create the mount point directory with the specified permissions, mount the disk, and update `/etc/fstab` for persistent mounting.
5. **Completion**: Once the process is complete, a success message will be displayed, and you can press `q` to quit.

### Example Usage

![example](example.gif)


## Error Handling

The program will display an error message if any step fails, such as if:

- The disk cannot be formatted.
- The mount point cannot be created.
- The disk cannot be mounted.
- There is an issue with updating `/etc/fstab`.

The error message will be shown in the `stateDone` step of the program.

## Commands

- **Enter**: Proceed with the selected action or input (e.g., format the disk, enter mount point).
- **q**: Quit the program.


## Contributing

Contributions are welcome! If you'd like to improve the utility, please fork the repository and create a pull request. Ensure that your code adheres to the project's coding standards and passes tests before submitting.

1. Fork the repository.
2. Create a new branch.
3. Make your changes.
4. Commit your changes with clear messages.
5. Open a pull request.
