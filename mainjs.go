package ziphttp

const (
	GUID = "ED93CDC6-E493-4DE8-971B-8BBCE66495FE"
	startPage = "webview.html"
	mainJS = `const electron = require('electron')
const app = electron.app
const BrowserWindow = electron.BrowserWindow
const globalShortcut = electron.globalShortcut
const session = electron.session
let mainWindow
function createWindow () {
  mainWindow = new BrowserWindow({width: 800, height: 600})
  mainWindow.on('closed', function () {
    mainWindow = null
  })
  const filter = {}
  session.defaultSession.webRequest.onBeforeSendHeaders(filter, (details, callback) => {
    details.requestHeaders['User-Agent'] = '%s'
    callback({cancel: false, requestHeaders: details.requestHeaders})
  })
  globalShortcut.register('CommandOrControl+C', () => {})
  mainWindow.loadURL('https://localhost:8080/%s')
}
app.commandLine.appendSwitch('ignore-certificate-errors')
app.on('ready', createWindow)
app.on('window-all-closed', function () {
  mainWindow.webContents.session.clearCache(function(){})
  app.quit()
})
`
	packageJson = `{
  "name": "Browser",
  "version": "1.0.0",
  "description": "A minimal Electron application",
  "main": "%s",
  "scripts": {
    "start": "electron ."
  },
  "repository": "",
  "keywords": [
    "Electron",
    "quick",
    "start",
    "tutorial",
    "demo"
  ],
  "author": "Peter Wang",
  "license": "CC0-1.0",
  "devDependencies": {
    "electron": "~1.6.2",
    "gulp": "^3.9.1",
    "gulp-sass": "^3.1.0"
  },
  "dependencies": {
    "electron": "^1.8.8",
    "favicon-getter": "^1.0.0",
    "jsonfile": "^2.4.0",
    "pdfjs-dist": "^2.0.489",
    "uuid": "^3.0.1"
  }
}
`
)