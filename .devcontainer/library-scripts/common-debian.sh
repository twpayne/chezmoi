#!/usr/bin/env bash
#-------------------------------------------------------------------------------------------------------------
# Copyright (c) Microsoft Corporation. All rights reserved.
# Licensed under the MIT License. See https://go.microsoft.com/fwlink/?linkid=2090316 for license information.
#-------------------------------------------------------------------------------------------------------------

# Syntax: ./common-debian.sh <install zsh flag> <username> <user UID> <user GID> <upgrade packages flag>

set -e

INSTALL_ZSH=${1:-"true"}
USERNAME=${2:-"$(awk -v val=1000 -F ":" '$3==val{print $1}' /etc/passwd)"}
USER_UID=${3:-1000}
USER_GID=${4:-1000}
UPGRADE_PACKAGES=${5:-"true"}

if [ "$(id -u)" -ne 0 ]; then
    echo 'Script must be run a root. Use sudo or set "USER root" before running the script.'
    exit 1
fi

# Treat a user name of "none" as root
if [ "${USERNAME}" = "none" ] || [ "${USERNAME}" = "root" ]; then
    USERNAME=root
    USER_UID=0
    USER_GID=0
fi

# Ensure apt is in non-interactive to avoid prompts
export DEBIAN_FRONTEND=noninteractive

# Install apt-utils to avoid debconf warning
apt-get -y install --no-install-recommends apt-utils 2> >( grep -v 'debconf: delaying package configuration, since apt-utils is not installed' >&2 )

# Get to latest versions of all packages
if [ "${UPGRADE_PACKAGES}" = "true" ]; then
    apt-get -y upgrade --no-install-recommends
fi

# Install common developer tools and dependencies
apt-get -y install --no-install-recommends \
    git \
    openssh-client \
    less \
    iproute2 \
    procps \
    curl \
    wget \
    unzip \
    nano \
    jq \
    lsb-release \
    ca-certificates \
    apt-transport-https \
    dialog \
    gnupg2 \
    libc6 \
    libgcc1 \
    libgssapi-krb5-2 \
    libicu[0-9][0-9] \
    liblttng-ust0 \
    libstdc++6 \
    zlib1g \
    locales

# Ensure at least the en_US.UTF-8 UTF-8 locale is available.
# Common need for both applications and things like the agnoster ZSH theme.
echo "en_US.UTF-8 UTF-8" >> /etc/locale.gen 
locale-gen

# Install libssl1.1 if available
if [[ ! -z $(apt-cache --names-only search ^libssl1.1$) ]]; then
    apt-get -y install  --no-install-recommends libssl1.1
fi
 
# Install appropriate version of libssl1.0.x if available
LIBSSL=$(dpkg-query -f '${db:Status-Abbrev}\t${binary:Package}\n' -W 'libssl1\.0\.?' 2>&1 || echo '')
if [ "$(echo "$LIBSSL" | grep -o 'libssl1\.0\.[0-9]:' | uniq | sort | wc -l)" -eq 0 ]; then
    if [[ ! -z $(apt-cache --names-only search ^libssl1.0.2$) ]]; then
        # Debian 9
        apt-get -y install  --no-install-recommends libssl1.0.2
    elif [[ ! -z $(apt-cache --names-only search ^libssl1.0.0$) ]]; then
        # Ubuntu 18.04, 16.04, earlier
        apt-get -y install  --no-install-recommends libssl1.0.0
    fi
fi

# Create or update a non-root user to match UID/GID - see https://aka.ms/vscode-remote/containers/non-root-user.
if id -u $USERNAME > /dev/null 2>&1; then
    # User exists, update if needed
    if [ "$USER_GID" != "$(id -G $USERNAME)" ]; then 
        groupmod --gid $USER_GID $USERNAME 
        usermod --gid $USER_GID $USERNAME
    fi
    if [ "$USER_UID" != "$(id -u $USERNAME)" ]; then 
        usermod --uid $USER_UID $USERNAME
    fi
else
    # Create user
    groupadd --gid $USER_GID $USERNAME
    useradd -s /bin/bash --uid $USER_UID --gid $USER_GID -m $USERNAME
fi

# Add add sudo support for non-root user
apt-get install -y sudo
echo $USERNAME ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/$USERNAME
chmod 0440 /etc/sudoers.d/$USERNAME

# Ensure ~/.local/bin is in the PATH for root and non-root users for bash. (zsh is later)
echo "export PATH=\$PATH:\$HOME/.local/bin" | tee -a /root/.bashrc >> /home/$USERNAME/.bashrc 
chown $USER_UID:$USER_GID /home/$USERNAME/.bashrc

# Optionally install and configure zsh
if [ "$INSTALL_ZSH" = "true" ] && [ ! -d "/root/.oh-my-zsh" ]; then 
    apt-get install -y zsh
    sh -c "$(curl -fsSL https://raw.githubusercontent.com/robbyrussell/oh-my-zsh/master/tools/install.sh)"
    echo "export PATH=\$PATH:\$HOME/.local/bin" >> /root/.zshrc
    cp -R /root/.oh-my-zsh /home/$USERNAME
    cp /root/.zshrc /home/$USERNAME
    sed -i -e "s/\/root\/.oh-my-zsh/\/home\/$USERNAME\/.oh-my-zsh/g" /home/$USERNAME/.zshrc
    chown -R $USER_UID:$USER_GID /home/$USERNAME/.oh-my-zsh /home/$USERNAME/.zshrc
fi

