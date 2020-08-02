# catbin

Anonymous encrypted pastebin

## Getting started

```sh
cd web
npm i
cp node_modules/purecss/build/pure-min.css .
cp node_modules/purecss/build/grids-responsive-min.css .
npm run build

cd ../api
go build
./catbin
# run as background process
# nohup ./catbin > log &
```