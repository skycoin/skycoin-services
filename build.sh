#!/usr/bin/env bash

# load env variables.
# shellcheck source=./build.conf
source "$(pwd)/build.conf"

ARCH=$(go env GOHOSTARCH)
OS=$(go env GOHOSTOS)

# Output directory.
FINAL_IMG_DIR=${ROOT}/skywireself/apps

# Download directories.
SKYWIRE_DIR=${ROOT}/bin

# for logging

info()
{
    printf '\033[0;32m[ INFO ]\033[0m %s\n' "${FUNCNAME[1]}: ${1}"
}

notice()
{
    printf '\033[0;34m[ NOTI ]\033[0m %s\n' "${FUNCNAME[1]}: ${1}"
}

warn()
{
    printf '\033[0;33m[ WARN ]\033[0m %s\n' "${FUNCNAME[1]}: ${1}"
}

error()
{
    printf '\033[0;31m[ ERRO ]\033[0m %s\n' "${FUNCNAME[1]}: ${1}"
}

create_folders()
{
    info "Creating output folder structure..."
    mkdir -p "$FINAL_IMG_DIR"
    mkdir -p "$SKYWIRE_DIR"

    info "Done!"
}


# Get skywire
get_skywire()
{
  local _DST=${SKYWIRE_DIR}/skywire.tar.gz # Download destination file name.

  if [ ! -f "${_DST}" ] ; then
    if [ ${OS} == darvin ] ; then
      notice "Downloading package from ${SKYWIRE_DARVIN_AMD64_DOWNLOAD_URL} to ${_DST}..."
      wget -c "${SKYWIRE_DARVIN_AMD64_DOWNLOAD_URL}" -O "${_DST}" || return 1
    elif [ ${OS} == linux ] ; then
      if [ ${ARCH} == amd64 ] ; then
        notice "Downloading package from ${SKYWIRE_AMD64_DOWNLOAD_URL} to ${_DST}..."
        wget -c "${SKYWIRE_AMD64_DOWNLOAD_URL}" -O "${_DST}" || return 1
      elif [ ${ARCH} == 386 ] ; then
        notice "Downloading package from ${SKYWIRE_386_DOWNLOAD_URL} to ${_DST}..."
        wget -c "${SKYWIRE_386_DOWNLOAD_URL}" -O "${_DST}" || return 1
      elif [ ${ARCH} == armv6 ] ; then
        notice "Downloading package from ${SKYWIRE_ARMV6_DOWNLOAD_URL} to ${_DST}..."
        wget -c "${SKYWIRE_ARMV6_DOWNLOAD_URL}" -O "${_DST}" || return 1
      elif [ ${ARCH} == armv7 ] ; then
        notice "Downloading package from ${SKYWIRE_ARMV7_DOWNLOAD_URL} to ${_DST}..."
        wget -c "${SKYWIRE_ARMV7_DOWNLOAD_URL}" -O "${_DST}" || return 1
      elif [ ${ARCH} == arm64 ] ; then
        notice "Downloading package from ${SKYWIRE_ARM64_DOWNLOAD_URL} to ${_DST}..."
        wget -c "${SKYWIRE_ARM64_DOWNLOAD_URL}" -O "${_DST}" || return 1
      fi
    fi
  else
      info "Reusing package in ${_DST}"
  fi

  info "Extracting package..."
  tar xvzf "${_DST}" -C "${SKYWIRE_DIR}" || return 1

  info "Copying..."
  cp -rf "$SKYWIRE_DIR"/apps/vpn-client "$FINAL_IMG_DIR" || return 1

  info "Copying..."
  cp -rf "$SKYWIRE_DIR"/skywire-cli "$FINAL_IMG_DIR" || return 1

  info "Cleaning..."
  rm -rf "${SKYWIRE_DIR}" || return 1

  info "Done!"
}


# Generate Config
gen_config(){
  info "Generating config..."
  # TODO: remove -t flag after prod starts working
  "$FINAL_IMG_DIR"/skywire-cli visor gen-config --is-hypervisor -r -t

  info "Copying..."
  mv -f ./skywire-config.json "${ROOT}"/skywireself || return 1
}

# main build block
main_build()
{
    # create output folder and it's structure
    create_folders || return 1

    # download resources
    get_skywire || return 1

    # generate config
    gen_config || return 1

    # all good signal
    info "Success!"
}

main_build || (error "Failed." && exit 1)
