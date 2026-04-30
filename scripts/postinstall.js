const fs = require('fs');
const path = require('path');
const https = require('https');
const { execSync } = require('child_process');

const REPO = 'Kukkerem/mcp-openrouter-search';
const BIN_DIR = path.join(__dirname, '..', 'bin');
const BIN_NAME = 'mcp-openrouter-search';

function getAssetName() {
  const platform = process.platform;
  const arch = process.arch;
  const osMap = { darwin: 'darwin', linux: 'linux', win32: 'windows' };
  const archMap = { x64: 'x86_64', arm64: 'aarch64' };
  const ext = platform === 'win32' ? '.exe' : '';
  const os = osMap[platform] || platform;
  const cpu = archMap[arch] || arch;
  return `${BIN_NAME}_${os}_${cpu}${ext}`;
}

function getLatestVersion() {
  try {
    const pkg = require('../package.json');
    return pkg.version;
  } catch {
    return 'v0.1.0';
  }
}

async function downloadBinary() {
  const version = getLatestVersion();
  const assetName = getAssetName();
  const url = `https://github.com/${REPO}/releases/download/${version}/${assetName}`;
  const binPath = path.join(BIN_DIR, BIN_NAME + (process.platform === 'win32' ? '.exe' : ''));

  if (fs.existsSync(binPath)) {
    console.log(`Binary already exists at ${binPath}`);
    return;
  }

  fs.mkdirSync(BIN_DIR, { recursive: true });

  console.log(`Downloading ${assetName}...`);

  return new Promise((resolve, reject) => {
    https.get(url, { headers: { 'User-Agent': 'npm' } }, (res) => {
      if (res.statusCode === 302 && res.headers.location) {
        https.get(res.headers.location, (res2) => {
          if (res2.statusCode !== 200) {
            reject(new Error(`Download failed: HTTP ${res2.statusCode}`));
            return;
          }
          const file = fs.createWriteStream(binPath);
          res2.pipe(file);
          file.on('finish', () => {
            file.close();
            fs.chmodSync(binPath, 0o755);
            console.log(`Binary installed at ${binPath}`);
            resolve();
          });
        }).on('error', reject);
      } else if (res.statusCode !== 200) {
        reject(new Error(`Download failed: HTTP ${res.statusCode}`));
      } else {
        const file = fs.createWriteStream(binPath);
        res.pipe(file);
        file.on('finish', () => {
          file.close();
          fs.chmodSync(binPath, 0o755);
          console.log(`Binary installed at ${binPath}`);
          resolve();
        });
      }
    }).on('error', reject);
  });
}

downloadBinary().catch((err) => {
  console.error('Failed to install binary:', err.message);
  console.error('You can manually download from https://github.com/' + REPO + '/releases');
  process.exit(0);
});
