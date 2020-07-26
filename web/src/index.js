import reload from "@riotjs/hot-reload";
import * as riot from "riot";
import App from "./app.riot";
import Encrypt from "./encrypt.riot";

riot.register("app", App)
riot.register("encrypt", Encrypt)
riot.mount("app")