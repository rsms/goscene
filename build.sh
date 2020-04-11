#!/bin/bash -e
cd "$(dirname "$0")"

BUILD_DIR=docs/app-dev

OPT_HELP=false
OPT_OPT=false
OPT_WATCH=false

# parse args
while [[ $# -gt 0 ]]; do
  case "$1" in
  -h*|--h*)
    OPT_HELP=true
    shift
    ;;
  -w*|--w*)
    OPT_WATCH=true
    shift
    ;;
  -O)
    OPT_OPT=true
    shift
    ;;
  *)
    echo "$0: Unknown option $1" >&2
    OPT_HELP=true
    shift
    ;;
  esac
done
if $OPT_HELP; then
  echo "Usage: $0 [options]"
  echo "Options:"
  echo "  -h   Show help."
  echo "  -O   Build optimized product."
  echo "  -w   Watch source files for changes and rebuild incrementally."
  exit 1
fi

# check esbuild
if ! (which esbuild >/dev/null); then
  esbuild_suffix=wasm
  if [[ "$OSTYPE" == "darwin"* ]]; then
    esbuild_suffix=darwin-64
  elif [[ "$OSTYPE" == "linux"* ]]; then
    esbuild_suffix=linux-64
  elif [[ "$OSTYPE" == "cygwin" ]] || \
       [[ "$OSTYPE" == "msys" ]] || \
       [[ "$OSTYPE" == "win32" ]] || \
       [[ "$OSTYPE" == "win64" ]]
  then
    esbuild_suffix=windows-64
  fi
  echo "esbuild not found in PATH. Please install with:" >&2
  echo "npm install -g esbuild-${esbuild_suffix}" >&2
  exit 1
fi

# check golib.js
if ! [ -f "docs/golib.js" ]; then
  echo "docs/golib.js is missing -- copy it from Go:" >&2
  echo 'cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" docs/golib.js' >&2
  exit 1
fi

# # check tinygo
# if ! (which tinygo >/dev/null); then
#   echo -n "tinygo not found in PATH. Please install" >&2
#   if [[ "$OSTYPE" == "darwin"* ]]; then
#     echo ":"
#     echo "  brew tap tinygo-org/tools"
#     echo "  brew install tinygo"
#   else
#     echo " from https://tinygo.org/"
#   fi

#   exit 1
# fi


mkdir -p "$BUILD_DIR"
BUILD_DIR_REL=$BUILD_DIR
pushd "$BUILD_DIR" >/dev/null
BUILD_DIR=$PWD
popd >/dev/null


WATCHFILE=$BUILD_DIR/.build.sh.watch

function fn_build_go {
  GO_SRCDIR=src
  pushd "$GO_SRCDIR" >/dev/null
  echo "go build $GO_SRCDIR -> $BUILD_DIR_REL/main.wasm"
  # tinygo build -o "$BUILD_DIR/main.wasm" -target wasm -no-debug .
  GOOS=js GOARCH=wasm go build -o "$BUILD_DIR/main.wasm"
  popd >/dev/null
}
function fn_build_js {
  if $OPT_OPT; then
    esbuild --define:DEBUG=false --bundle --sourcemap --minify \
      "--outfile=$BUILD_DIR/host.js" src/host.js
  else
    esbuild --define:DEBUG=true --bundle --sourcemap \
      "--outfile=$BUILD_DIR/host.js" src/host.js
  fi
}

function fn_watch_go {
  while true; do
    fswatch -1 -l 0.2 -r -E --exclude='.+' --include='\.go$' src >/dev/null
    if ! [ -f "$WATCHFILE" ] || [ "$(cat "$WATCHFILE")" != "y" ]; then break; fi
    set +e ; fn_build_go ; set -e
  done
}

function fn_watch_js {
  while true; do
    fswatch -1 -l 0.2 -r -E --exclude='.+' --include='\.js$' src >/dev/null
    if ! [ -f "$WATCHFILE" ] || [ "$(cat "$WATCHFILE")" != "y" ]; then break; fi
    set +e ; fn_build_js ; set -e
  done
}


fn_build_go &
fn_build_js &

if $OPT_WATCH; then
  echo y > "$WATCHFILE"

  # make sure we can ctrl-c in the while loop
  function fn_stop {
    echo n > "$WATCHFILE"
    exit
  }
  trap fn_stop SIGINT

  # make sure background processes are killed when this script is stopped
  pids=()
  function fn_cleanup {
    set +e
    for pid in "${pids[@]}"; do
      kill $pid 2>/dev/null
      wait $pid
      kill -9 $pid 2>/dev/null
      echo n > "$WATCHFILE"
    done
    set -e
  }
  trap fn_cleanup EXIT

  # wait for initial build
  wait

  # start web server
  if (which serve-http >/dev/null); then
    serve-http -p 3891 -quiet docs &
    pids+=( $! )
    echo "Web server listening at http://localhost:3891/"
  else
    echo "Tip: Install serve-http to have a web server run."
    echo "     npm install -g serve-http"
  fi

  echo "Watching source files for changes..."

  fn_watch_go &
  pids+=( $! )

  fn_watch_js
else
  wait
fi
