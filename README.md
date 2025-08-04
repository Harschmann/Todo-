# todoplusplus 🚀

A TUI (Terminal User Interface) to track your competitive programming grind, built with Go.

![Screenshot](path/to/your/screenshot.png) ## Features

* **Add, Edit, Delete & View Logs:** Full CRUD functionality for your problem logs.
* **TUI Form:** A fast, multi-view terminal interface for data entry built with Bubbletea.
* **Local Database:** All data is stored persistently in a local BoltDB database.
* **Filtering:** The "View Logs" screen has a real-time filter to search by Question ID, Platform, Topic, or Difficulty.
* **Google Calendar Sync:** Automatically creates and deletes a corresponding event on your Google Calendar for every log entry.
* **Automated Reminders:** A background service sends a reminder email via the Gmail API on days you haven't solved a problem.
* **Periodic Backups:** Automatically creates a JSON backup of your database every 30 minutes, with rotation to save the last 150 copies.
* **Excel Export:** Export all your logs to an `.xlsx` file with a command-line flag.

## Installation

#### macOS / Linux (via Homebrew)
```sh
brew install Harschmann/tap/todoplusplus