# catbin

Anonymous encrypted pastebin

https://github.com/ryyan/catbin/assets/4228816/35f1689e-d072-46d8-bb80-ae7f53d5cfb0

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
