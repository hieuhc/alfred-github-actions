# Alfred workflow for Github Actions

A shortcut to work with Github Actions(GHA) using Alfred. Users can register to watch for a GHA run, 
and Alfred will send a notifcation when the run is done.

## Install
- Install [Alfred powerpack](https://www.alfredapp.com/powerpack/).
- Download the latest version [here](https://github.com/hieuhc/alfred-github-actions/releases). 
- Install the workflow to Alfred.

## Usage
`gha-login <token>`: Authenticate with Github [personal access token](https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/creating-a-personal-access-token).
This token will be stored in the user's keychain at local. Users will probably be asked to enter their "login" password to access the keychain when logging in the first time, click on "Always Allow" to avoid having to re-enter the keychain password next times.


`gha-refresh`: To update list of repositories.

Use <kbd>↩</kbd>  to navigate the path to a GHA run. This path follows _respository_, _workflow_ and _run_.
- Standing at the _repository_:
    - <kbd>↩</kbd> to see a list of workflows that belong to this respository.
    - <kbd>CMD</kbd><kbd>↩</kbd> to open the web url of that repo.
- Standing at the _workflow_:
    - <kbd>↩</kbd> to see a list of runs triggered for that workflow.
    - <kbd>CMD</kbd><kbd>↩</kbd> to open the web url of that workflow.
- Standing at the _run_:
    - <kbd>↩</kbd> to open the web url of that run.
    - <kbd>CMD</kbd><kbd>↩</kbd> to register the run to watch. Alfred will send a notification message when the run is done.


