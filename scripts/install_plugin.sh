#!/usr/bin/env bash
#
# install_plugins.sh
#
# This script installs the plugin in ~/.terraform.d/plugins

set -e

os="$(uname | tr '[:upper:]' '[:lower:]')"
arch="$(uname -m)"
plugins_dir="${HOME}/.terraform.d/plugins"

install_plugin() {
  plugin=$1
  version=0.0.1
  plugin_name=terraform-provider-$(basename "${plugin}")
  plugin_location="${GOBIN:-${GOPATH}/bin}/${plugin_name}"
  echo "Installing Terraform plugin ${plugin}..."
  file="${plugin_name}_v${version}-${os}-${arch}"
  plugin_dst="${plugins_dir}/${plugin}/${version}/${os}_${arch}/${file}"
  mkdir -p "$(dirname "${plugin_dst}")"
  echo "location: ${plugin_location}"
  cp "${plugin_location}" "${plugin_dst}"
  echo "Copied to ${plugin_dst}"
}

install_plugin "terraform.local.com/TrueLayer/slack"
