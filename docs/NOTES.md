# Otto: Notes

## Overview

Otto is a web server application designed to handle the upload, parsing, and generation of maps for TribeNet turn reports. Initially, Otto assumed that each user would submit reports for their clan. However, the application is being updated to support sectioning reports, allowing users to submit reports for their clan and their allies.

## Report Identification

To accommodate the new sectioning feature, the process of identifying reports needs to be updated. The key for the report table will be an integer created by a sequence generator. The user-facing identifier for a report will be a combination of the turn year, turn month, report clan number, and the user's clan number. For example: `0900-01.0138-0138` (TurnYear-TurnMonth.ReportClanNumber.UserClanNumber). Users will rely on the timestamp to determine the latest version of a report.

## Report Versioning

There are a few considerations regarding report versioning:
- Should we allow more than one report version per user?
- Should we limit the number of versions for any report?
- Should users be responsible for purging prior versions?
- Or should we wait until they're ready to generate a map and require them to purge prior versions before proceeding?

## Upload Process

### Duplicate Check

Upon upload, Otto calculates a checksum for the file and checks for duplicates. If a duplicate is found, Otto will not upload the file, and the user will be notified with a link to the existing report.

### Sectioning

Otto splits the file into sections, one section per element. If there are errors during sectioning, Otto assumes the file is not a valid turn report, notifies the user of the errors, and fails the upload. If the first record is not the clan element, Otto reports an error and does not update the database.

For each valid section:
1. The element header is written to the `reports_sections` table.
2. Otto verifies that the element rolls up to the clan. If not, an error is recorded in the `reports_sections` record, and Otto skips to the next section.
3. If the element is valid, all section lines are written to the `reports_lines` table.

Upon completion, the user is notified and provided a link to the new report for viewing and parsing.

## Parsing

### Turn Reports

- Turn reports may contain typos from the production process; these need to be fixed by the user.
- Otto parses the report and displays any errors to the user.
- Users should parse each report before generating a map.
- Otto parses each section, checking each line for errors. All errors are saved as a message on the corresponding line.
- Otto stops parsing a line on the first error but tries to parse all lines in each section.
- After parsing, users should review the errors, correct the turn report, and upload the corrected file.

## Map Generation

To generate a map, users must navigate to their turns page, which displays the available turns (culled from the list of uploaded reports). When a turn is selected, Otto checks for multiple versions of reports. If multiple versions exist, the user must manually delete older versions before proceeding.

Assuming all prerequisites are met, the user can generate the map. The generator parses the reports, logging any errors. If the user parsed the reports before generating the map, there should be no errors. If no errors are found, Otto creates a map file and immediately downloads it to the user's browser. Maps are not stored on the server.

## Database Structure

Otto uses four tables to store turn reports in the database:
1. `uploads` - Stores metadata for the upload (file name, checksum, timestamp, unique id).
2. `reports` - Stores the report metadata, including the user-facing identifier and timestamp.
3. `reports_sections` - Stores metadata for section of the report, including the element header and any errors.
4. `reports_lines` - Stores the individual lines of each section, along with any parsing errors.

