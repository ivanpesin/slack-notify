# Slack notify

This tool implements the communication with slack for various purposes. It
supports the following 3 modes:

- `msg`: sends a short message to slack. Useful for notifications from scripts.
- `monit`: sends a *monit* event to slack. See below the example of
  configuration and the message
- `mail`: forwards local mail to slack
