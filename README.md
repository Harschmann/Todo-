# todoplusplus 🚀

Tired of clunky spreadsheets or plain text files to track your competitive programming grind? **`todoplusplus`** is a blazingly fast, terminal-based application built in Go, designed to manage your entire CP journey without ever leaving the keyboard.

It's local-first, powerful, and integrates with the tools you already use.

---

## 🎥 Demo

*(Placeholder for your awesome feature demo video!)*

---

## ✨ Features

* **💻 Sleek Terminal UI:** A fast, multi-view terminal interface for all your actions, built with the powerful Bubbletea framework.
* **💾 Local-First Database:** All your data is stored securely and instantly on your own machine in an embedded BoltDB database. No cloud dependency for your core data.
* **⚙️ Full CRUD Functionality:** Add, view, edit, and delete your problem logs with a seamless, intuitive workflow.
* **🔍 Real-time Filtering:** Instantly search through hundreds of logs by Question ID, Platform, Topic, or Difficulty in the "View Logs" screen.
* **🗓️ Google Calendar Sync:** Automatically creates and deletes a corresponding event on your Google Calendar for every log entry, giving you a powerful visual overview of your consistency.
* **📧 Smart Email Reminders:** A background service can be configured to send a fun, randomized reminder via the Gmail API on days you forget to practice.
* **📦 Automated Backups:** Your database is automatically backed up periodically, with smart rotation to save the last 150 copies, ensuring your data is always safe.
* **📄 Excel Export:** Export all your logs to a clean `.xlsx` spreadsheet with a single command.

---

## 🛠️ Installation

#### macOS / Linux (via Homebrew)

This is the recommended method for macOS and Linux users.

1.  **Tap the repository** (only needs to be done once):
    ```sh
    brew tap Harschmann/homebrew-tap
    ```
2.  **Install the app:**
    ```sh
    brew install todoplusplus
    ```
    To upgrade to the latest version in the future, just run `brew upgrade todoplusplus`.

#### Windows / Other Systems (from GitHub Releases)

1.  Go to the [**Releases**](https://github.com/Harschmann/todoplusplus/releases) page on GitHub.
2.  Download the latest `.zip` (for Windows) or `.tar.gz` (for Linux) file for your system's architecture (usually `amd64`).
3.  Unzip the file and place the `todoplusplus` executable somewhere in your system's `PATH`.

---

## 🚀 Usage

#### Main Application
To start the main Terminal User Interface, simply run:
```sh
todoplusplus
