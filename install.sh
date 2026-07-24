#!/bin/bash
# Statping installation script for Linux, Mac, and Windows.
set -e

reset="\033[0m"
red="\033[31m"
green="\033[32m"
yellow="\033[33m"
cyan="\033[36m"
white="\033[37m"
repo="https://github.com/statping-ng/statping-ng"

statping_get_tarball() {
  local os_type="$1"
  local arch_type="$2"
  fext='tar.gz'
  if [ "${os_type}" = 'windows' ]; then
    fext='zip'
    arch_type='x64'
  fi
  url="${repo}/releases/latest/download/statping-${os_type}-${arch_type}.${fext}"
  printf "%b> Downloading latest version for %s %s...\n%s %b\n" "$cyan" "$os_type" "$arch_type" "$url" "$reset"
  tarball_tmp=$(mktemp -t statping.tar.gz.XXXXXXXXXX)
  if curl --fail -L -s -o "$tarball_tmp" "$url"; then
    temp=$(mktemp -d statping.XXXXXXXXXX)
    if [ "${os_type}" = 'windows' ]; then
      unzip "$tarball_tmp" -d "$temp"
    else
      tar xzf "$tarball_tmp" -C "$temp"
    fi
    printf "%b> Installing to %s/statping-ng\n" "$green" "$DEST"
    mv "$temp"/statping "$DEST"
    rm -rf "$temp"
    rm -f "${tarball_tmp:?}"*
    printf "%b> Statping-ng is now installed! %b\n" "$cyan" "$reset"
    printf "%b>   Repo:     %s %b\n" "$white" "$repo" "$reset"
    printf "%b>   Wiki:     %s/wiki %b\n" "$white" "$repo" "$reset"
    printf "%b>   Issues:   %s/issues %b\n" "$white" "$repo" "$reset"
    printf "%b> Try to run \"statping help\" %b\n" "$cyan" "$reset"
  else
    printf "%b> Failed to download %s.%b\n" "$red" "$url" "$reset"
    exit 1
  fi
}

statping_reset() {
  unset -f statping_install statping_reset statping_get_tarball statping_verify_or_quit statping_brew_install getOS getArch
}

statping_brew_install() {
  if command -v brew > /dev/null 2>&1; then
    printf "%bUsing Brew to install!%b\n" "$white" "$reset"
    printf "%b---> brew tap statping-ng/statping-ng%b\n" "$yellow" "$reset"
    brew tap statping-ng/statping-ng
    printf "%b---> brew install statping%b\n" "$yellow" "$reset"
    brew install statping-ng
    printf "%bBrew installation is complete!%b\n" "$green" "$reset"
    printf "%bYou can use 'brew upgrade' to upgrade Statping next time.%b\n" "$yellow" "$reset"
  else
    statping_get_tarball "$OS" "$ARCH"
  fi
}

statping_install() {
  printf "%bInstalling Statping-ng!%b\n" "$white" "$reset"
  getOS
  getArch
  statping_get_tarball "$OS" "$ARCH"
  statping_reset
}

statping_verify_or_quit() {
  read -r -p "$1 [y/N] " REPLY
  echo
  if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    printf "%b> Aborting%b\n" "$red" "$reset"
    exit 1
  fi
}

getOS() {
  OS="$(uname)"
  case $OS in
    'Linux')
      OS='linux'
      DEST=/usr/local/bin
      ;;
    'FreeBSD')
      OS='freebsd'
      DEST=/usr/local/bin
      ;;
    'OpenBSD')
      OS='openbsd'
      DEST=/usr/local/bin
      ;;
    'WindowsNT' | 'MINGW*' | 'CYGWIN*')
      OS='windows'
      DEST=/usr/local/bin
      ;;
    'Darwin')
      OS='darwin'
      DEST=/usr/local/bin
      ;;
    'SunOS')
      OS='linux'
      DEST=/usr/local/bin
      ;;
    *) ;;
  esac
}

getArch() {
  MACHINE_TYPE=$(uname -m)
  if [ "${MACHINE_TYPE}" = 'x86_64' ]; then
    ARCH="amd64"
  elif [ "${MACHINE_TYPE}" = 'arm' ]; then
    ARCH="arm"
  elif [ "${MACHINE_TYPE}" = 'arm64' ] || [ "${MACHINE_TYPE}" = 'aarch64' ]; then
    ARCH="arm64"
  else
    ARCH="386"
  fi
}

cd ~ || exit 1
statping_install "$1" "$2"
