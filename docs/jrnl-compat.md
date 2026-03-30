# jrnl Compatibility

> **Note:** This document is a work in progress. It will be completed after all feature-parity passes are done.

## Backend

jrnl supports multiple storage backends: DayOne, encrypted single-file, and folder-of-files. **jrnl-md implements only the folder-based markdown backend.** Journal files are stored as one Markdown file per day at `YYYY/MM/DD.md`.

There is no `type` config key and no plan to add one. If you are migrating from a jrnl DayOne or single-file journal, you will need to export to text first and then import into jrnl-md.

## What jrnl-md Does and Does Not Implement

*(Full compatibility matrix to be documented here after all feature passes are complete.)*
