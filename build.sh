#!/usr/bin/env bash

# This is the main script to build the Skybian OS for Skycoin Official and Raspberry Pi miners.
#
# Author: evanlinjin@github.com, @evanlinjin in Telegram
# Modified by: asxtree@github.com, @asxtree in Telegram
# Skycoin / Rudi team
#

# load env variables.
# shellcheck source=./build.conf
source "$(pwd)/build.conf"

## Variables.

# Needed tools to run this script, space separated
# On arch/manjaro, the qemu-aarch64-static dependency is satisfied by installing the 'qemu-arm-static' AUR package.
NEEDED_TOOLS="rsync wget 7z cut awk sha256sum gzip tar e2fsck losetup resize2fs truncate sfdisk qemu-aarch64-static qemu-arm-static go"

# Check if build variables were set
# If the BOARD and ARCH variables are not set in the build command, it will build a skybian image for Orange Pi Prime by default
if [ -z ${BOARD} ] || [ -z ${ARCH} ] ; then
  BOARD=prime
  ARCH=arm64
fi


# Output directory.
FINAL_IMG_DIR=${ROOT}/skywireself/apps

# Download directories.
SKYWIRE_DIR=${ROOT}/bin



##############################################################################
# This bash file is structured as functions with specific tasks, to see the
# tasks flow and comments go to bottom of the file and look for the 'main'
# function to see how they integrate to do the  whole job.
##############################################################################

# Capturing arguments to show help
if [ "$1" == "-h" ] || [ "$1" == "--help" ] ; then
    # show help
    cat << EOF

$0, Skybian build script.

This script builds the Skybian base OS to be used on the Skycoin
official Skyminers, there is just a few parameters:

-h / --help     Show this help
-p              Pack the image and checksums in a form ready to
                deploy into a release. WARNING for this to work
                you need to run the script with no parameters
                first
-c              Clean everything (in case of failure)

No parameters means image creation without checksum and packing

To know more about the script work, please refers to the file
called Building_Skybian.md on this folder.

Latest code can be found on https://github.com/skycoin/skybian

EOF

    exit 0
fi

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

# Build the output/work folder structure, this is excluded from the git
# tracking on purpose: this will generate GB of data on each push
create_folders()
{
    # output-x [main folder]
    #   /final [this will be the final images dir]
    #   /parts [all thing we download from the internet]
    #   /mnt [to mount resources, like img fs, etc]
    #   /image [all image processing goes here]

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
    if [ ${BOARD} == rpi ] ; then
      notice "Downloading package from ${SKYWIRE_ARMV7_DOWNLOAD_URL} to ${_DST}..."
      wget -c "${SKYWIRE_ARMV7_DOWNLOAD_URL}" -O "${_DST}" || return 1
    elif [ ${BOARD} == rpiw ] ; then
      notice "Downloading package from ${SKYWIRE_ARMV6_DOWNLOAD_URL} to ${_DST}..."
      wget -c "${SKYWIRE_ARMV6_DOWNLOAD_URL}" -O "${_DST}" || return 1
    else
      notice "Downloading package from ${SKYWIRE_ARM64_DOWNLOAD_URL} to ${_DST}..."
      wget -c "${SKYWIRE_ARM64_DOWNLOAD_URL}" -O "${_DST}" || return 1
    fi
  else
      info "Reusing package in ${_DST}"
  fi

  info "Extracting package..."
  tar xvzf "${_DST}" -C "${SKYWIRE_DIR}" || return 1

  info "Copying..."
  cp -rf "$SKYWIRE_DIR"/apps/vpn-client "$FINAL_IMG_DIR" || return 1

  info "Cleaning..."
  rm -rf "${SKYWIRE_DIR}" || return 1

  info "Done!"
}


# calculate md5, sha1 and compress
calc_sums_compress()
{
	  FINAL_IMG_DIR="${ROOT}/output/${BOARD}/final"
  
  # change to final dest
    cd "${FINAL_IMG_DIR}" ||
      (error "Failed to cd." && return 1)

  # info
    info "Calculating the md5sum for the image, this may take a while"

  # cycle for each one
    for img in $(find -- *.img -maxdepth 1 -print0 | xargs --null) ; do
    # MD5
      info "MD5 Sum for image: $img"
      md5sum -b "${img}" > "${img}.md5"

    # sha1
      info "SHA1 Sum for image: $img"
      sha1sum -b "${img}" > "${img}.sha1"

    # compress
      info "Compressing, this will take a while..."
      name=$(echo "${img}" | rev | cut -d '.' -f 2- | rev)
      tar -cvzf "${name}.tar.gz" "${img}"*
    done

    cd "${ROOT}" || return 1
    info "Done!"
}

clean_image()
{
  sudo umount "${FS_MNT_POINT}/sys"
  sudo umount "${FS_MNT_POINT}/proc"
  sudo umount "${FS_MNT_POINT}/dev/pts"
  sudo umount "${FS_MNT_POINT}/dev"

  sudo sync
  sudo umount "${FS_MNT_POINT}"

  sudo sync
  # only do so if IMG_LOOP is set
  [[ -n "${IMG_LOOP}" ]] && sudo losetup -d "${IMG_LOOP}"
}

clean_output_dir()
{
  cd "${PARTS_OS_DIR}" && find . -type f ! -name '*.xz' -delete
  cd "${PARTS_SKYWIRE_DIR}" && find . -type f ! -name '*.tar.gz' -delete && rm -rf bin
  # cd "${FINAL_IMG_DIR}" && find . -type f ! -name '*.tar.gz' -delete
  rm -v "${BASE_IMG}"
  
  # cd to root.
  cd "${ROOT}" || return 1
}


# main build block
main_build()
{

    # create output folder and it's structure
    create_folders || return 1

    # download resources
    get_skywire || return 1

    # all good signal
    info "Success!"
}

main_clean()
{
  clean_output_dir
  clean_image || return 0
}

# clean exec block
main_package()
{
    tool_test || return 1
    #create_folders || return 1
    calc_sums_compress || return 1
    info "Success!"
}

case "$1" in
"-p")
    # Package image.
    main_package || (error "Failed." && exit 1)
    ;;
"-c")
    # Clean in case of failures.
    main_clean || (error "Failed." && exit 1)
    ;;
*)
    main_build || (error "Failed." && exit 1)
    ;;
 esac
