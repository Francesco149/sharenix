#!/bin/sh

git pull origin master
go get ./...

echo ""
echo "Compiling and Stripping"
LDFLAGS="$LDFLAGS -static -s -w -no-pie -Wl,--gc-sections"
LDFLAGS="$LDFLAGS $(pkg-config --static --libs gtk+-2.0)"
go build --ldflags "-linkmode external -extldflags '$LDFLAGS'"

echo ""
echo "Packaging"
folder="sharenix-$(uname -m)"
rm -rf "$folder"
mkdir -p "$folder"
mv ./sharenix $folder/sharenix
cp ./sharenix.json $folder/sharenix.json
git archive HEAD --prefix=src/ -o "$folder"/src.tar
cd "$folder"
tar xf src.tar
cd ..

rm "$folder".tar.xz
tar -cvJf "$folder".tar.xz \
    "$folder"/sharenix \
    "$folder"/sharenix.json \
    "$folder"/src

echo ""
echo "Result:"
tar tf "$folder".tar.xz

readelf --dynamic "$folder"/sharenix || exit 1
ldd "$folder"/sharenix && exit 1

exit 0

