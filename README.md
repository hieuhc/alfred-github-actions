# Alfred workflow for Github Actions

A shortcut to work with Github Actions workflows and runs using Alfred. Users can register to watch for a GHA run, 
and Alfred will send a notifcation when the run is finished.

## Install
- Install [Alfred powerpack](https://www.alfredapp.com/powerpack/).
- Download the latest version produced [here](). TODO update link here
- Install the workflow to Alfred.

## Usage
`gha-login <token>`: Authenticate with Github [personal access token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token).
This token is stored in the user's keychain on local. You will be asked to enter your "login" password to access the keychain when logging in the first time, click on "Always Allow" so that
you don't have to re-enter the keychain password next times.


`gha-refresh`: To update list of repositories.

Use <kbd>↩</kbd>  to navigate the path to a GHA run. This path follows `respository`, `workflow` and `run`.
- Standing at the `repository`:
    - <kbd>↩</kbd> to see a list of workflows that belong to this respository.
    - <kbd>CMD</kbd><kbd>↩</kbd> to open the web url of that repo.
- Standing at the `workflow`:
    - <kbd>↩</kbd> to see a list of runs triggered for that workflow.
    - <kbd>CMD</kbd><kbd>↩</kbd> to open the web url of that workflow.
- Standing at the `run`:
    - <kbd>↩</kbd> to open the web url of that run.
    - <kbd>CMD</kbd><kbd>↩</kbd> to register the run to watch. Alfred will send a notification message when the run is done.


