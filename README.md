# wrikemeup

Never gonna give you up, never gonna let you down... but it will log your hours into Wrike!

# Setup

## 1. Fork this repostitory

## 2. Setup the users

Wrike me up! expects two secrets as environment variables. 

- `USERS` in a base64 encoded JSON format. 

    The JSON must be an array of objects, each containing the following keys:
    
    ```json
    [
        {
            "github_username": "rick",
            "wrike_email": "rick@wrikemeup.com",
            "wrike_token": "someLongWrikeToken"
        },
        {
            "github_username": "roll",
            "wrike_email": "roll@wrikemeup.com",
            "wrike_token": "otherLongWrikeToken"
        }
    ]
    ```

## 3. `BOT_TOKEN` 

Wrike me up! bot will answer to your commands sending a comment to the issue.

For this to work, you need to set the `BOT_TOKEN` environment variable with a GitHub token with `repo` permissions.

