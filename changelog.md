# ChangeLog

## 0.0.1
### description
Initial release

### Added
- Initiial cli
- Wizard for selecting vault
- New vault type: Azure Key Vault

## 0.0.2
### description
major rewrite as the current version left a lot to be desired

### added

- completley new init flow and ability to controll environments
- added methods for mulit-vault pull
- reworte config to be able to manage relationship between local secret and vault value
- suggested a new way to manage local secret files for easier gitignore
- rewote how environment was handled. its not a direct part of the cli, and not a argument
  - polyenv !{env} <command>
- way better tui for the wizard
