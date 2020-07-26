import reload from '@riotjs/hot-reload'
import * as riot from 'riot'
import App from './app.riot'

riot.register('app', App)
riot.mount('app')