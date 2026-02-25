# This file will be copied to the devcontainer's .bashrc

# Define command aliases
alias c='clear'
alias t='task'
alias ll='ls -alh'
alias ..='cd ..'

# Set the user file-creation mode mask to 000, which allows all
# users read, write, and execute permissions for newly created files.
umask 000

# Run the 'fixperms' task that fixes the permissions of the files and
# directories in the project.
chmod -R 777 /workspaces

# Configure Git to ignore ownership and file mode changes.
git config --global --add safe.directory '*'
git config core.fileMode false
