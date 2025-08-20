Hide 
#generate polyenv file
Type "polyenv init dev --vault devvault --accept-default-settings"
Enter
Sleep 50ms

Enter

Type "dev-vault"
Enter

Sleep 50ms
Enter


Type cls
Enter
Show



Type "polyenv status"
Enter
Sleep 1s

# Type a command in the terminal.
Type "polyenv !dev add secret"
Sleep 1s
Enter

Sleep 1s
Enter

#select secrets
Space
Sleep 1s
Down
Space
Sleep 1s

Enter
Type "local-env-key"
Sleep 1s
Enter

Type "my_other_secret"
Sleep 1s
Enter


Sleep 500ms

Type "polyenv status"
Enter
Sleep 1s



# # Down
# Sleep 500ms
# Enter


# Sleep 500ms
# Down
# Sleep 500ms
# Enter

# Sleep 500ms
# Type "dev-vault"
# Sleep 500ms
# Enter

# Enter

# Sleep 1
# Type gc "dev.polyenv.toml"
# Enter
# Sleep 2

#select "New"
# Down
# Sleep 500ms
# Enter

# #give env a name
# Type Demo
# Sleep 2s
# Enter

# #keep defaults
# Sleep 2s
# Enter

# #Yes to add vault
# Sleep 1s
# Left
# Sleep 1s
# Enter

# #select a dev-vault
# Down
# Sleep 1s
# Enter

# #select vault
# Sleep 2s
# Enter

# #enter vault name
# Type my-dev-vault
# Sleep 2s
# Enter

# #yes to add secret
# Sleep 500ms
# Left
# Enter

Sleep 2s
# Sleep 2s


Ctrl+C
# # Run the command by pressing enter.
# Enter

# # Admire the output for a bit.
# Sleep 5s