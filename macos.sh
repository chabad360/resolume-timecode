#!/bin/bash
# Copyright (c) 2017 David Pennington (modified by Mendel Greenberg) LICENSE: MIT

# Mac OSX .app builder

function die {
	echo "ERROR: $1" > /dev/null 1>&2
	exit 1
}

if [ "$#" -ne 5 ]; then
	die "Usage: `basename $0` AppNameHere icon-file.svg com.example.app 0 1"
fi

APPNAME=$1
ICONNAME=$2
APPID=$3
MAJOR=$4
MINOR=$5

if [ ! -f $ICONNAME ]; then
	die "Image file for icon not found"
fi

mkdir -p "$APPNAME.app/Contents/"{MacOS,Resources}

cat > "$APPNAME.app/Contents/Info.plist" <<END
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple Computer//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>CFBundleGetInfoString</key>
  <string>$APPNAME</string>
  <key>CFBundleExecutable</key>
  <string>$APPNAME</string>
  <key>CFBundleIdentifier</key>
  <string>$APPID</string>
  <key>CFBundleName</key>
  <string>$APPNAME</string>
  <key>CFBundleIconFile</key>
  <string>icon.icns</string>
  <key>CFBundleShortVersionString</key>
  <string>${MAJOR}.${MINOR}</string>
  <key>CFBundleInfoDictionaryVersion</key>
  <string>6.0</string>
  <key>CFBundlePackageType</key>
  <string>APPL</string>
  <key>IFMajorVersion</key>
  <integer>$MAJOR</integer>
  <key>IFMinorVersion</key>
  <integer>$MINOR</integer>
  <key>NSHighResolutionCapable</key><true/>
  <key>NSSupportsAutomaticGraphicsSwitching</key><true/>
</dict>
</plist>
END

cp $ICONNAME "$APPNAME.app/Contents/Resources/"
cd "$APPNAME.app/Contents/Resources/"

fileName="$(basename $ICONNAME)"
postfix=${fileName##*.}

if [[ $postfix == 'svg' ]]; then
    qlmanage -z -t -s 1024 -o ./ "$fileName"
    fileName=${fileName}.png
fi

echo $fileName

mkdir icon.iconset

rsvg-convert -z 16 "$fileName" > icon.iconset/icon_16x16.png
rsvg-convert -h 32 "$fileName" > icon.iconset/icon_16x16@2x.png
cp icon.iconset/icon_16x16@2x.png icon.iconset/icon_32x32.png
rsvg-convert -h 64 "$fileName" > icon.iconset/icon_32x32@2x.png
rsvg-convert -h 128 "$fileName" > icon.iconset/icon_128x128.png
rsvg-convert -h 256 "$fileName" > icon.iconset/icon_128x128@2x.png
cp icon.iconset/icon_128x128@2x.png icon.iconset/icon_256x256.png
rsvg-convert -h 512 "$fileName" > icon.iconset/icon_256x256@2x.png
cp icon.iconset/icon_256x256@2x.png icon.iconset/icon_512x512.png
rsvg-convert -h 1024 "$fileName" > icon.iconset/icon_512x512@2x.png

# Create .icns file
iconutil -c icns icon.iconset

# Cleanup
#rm -R icon.iconset
#rm $fileName
