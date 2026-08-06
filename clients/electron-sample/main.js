const {app, BrowserWindow} = require('electron');
app.commandLine.appendSwitch('ignore-certificate-errors');
app.whenReady().then(async () => {
  const win = new BrowserWindow({ show: false, webPreferences: { sandbox: false } });
  try {
    await win.loadURL('https://electron-sample.fp.lab.local:8443/');
    console.log('electron-sample loaded');
  } catch (e) {
    console.error('load error (CH may still be saved)', e && e.message ? e.message : e);
  }
  setTimeout(() => app.quit(), 2000);
});
