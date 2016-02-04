#/bin/bash
set -e
cd "$( dirname "${BASH_SOURCE[0]}" )/.."

if hash protoc 2>/dev/null; then
	exit 0
fi

echo "Installing protoc"

DIR=/tmp/protoc-v3.0.0-beta-2
VER=3.0.0-beta-2
ZIP=${DIR}/${VER}.zip

rm -rf ${DIR}
brew install autoconf automake libtool
mkdir -p ${DIR}
curl -L https://github.com/google/protobuf/archive/v${VER}.zip -o ${ZIP}
unzip ${ZIP} -d ${DIR}

cd ${DIR}/protobuf-${VER}
./autogen.sh
./configure
make
make check
sudo make install
