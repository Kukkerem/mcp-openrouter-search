const fs = require('fs');
const path = require('path');
const https = require('https');
const { execSync } = require('child_process');

const REPO = 'Kukkerem/mcp-openrouter-search';
const BIN_DIR = path.join(__dirname, '..', 'bin');
const BIN_NAME = 'mcp-openrouter-search';

function getPlatform() {
  const osMap = { darwin: 'darwin', linux: 'linux', win32: 'windows' };
  const archMap = { x64: 'x86_64', arm64: 'aarch64' };
  const os = osMap[process.platform] || process.platform;
  const arch = archMap[process.arch] || process.arch;
  const ext = process.platform === 'win32' ? '.zip' : '.tar.gz';
  return { os, arch, ext };
}

function getAssetName() {
  const { os, arch, ext } = getPlatform();
  return `${BIN_NAME}_${os}_${arch}${ext}`;
}

function getLatestVersion() {
  try {
    const pkg = require('../package.json');
    const v = pkg.version;
    return v.startsWith('v') ? v : `v${v}`;
  } catch {
    return 'v0.1.0';
  }
}

function downloadFile(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);
    https.get(url, { headers: { 'User-Agent': 'npm' } }, (res) => {
      if (res.statusCode === 302 && res.headers.location) {
        https.get(res.headers.location, (res2) => {
          if (res2.statusCode !== 200) {
            reject(new Error(`HTTP ${res2.statusCode}`));
            return;
          }
          res2.pipe(file);
          file.on('finish', () => { file.close(); resolve(); });
        }).on('error', reject);
      } else if (res.statusCode !== 200) {
        reject(new Error(`HTTP ${res.statusCode}`));
      } else {
        res.pipe(file);
        file.on('finish', () => { file.close(); resolve(); });
      }
    }).on('error', reject);
  });
}

function extractArchive(archivePath, extractDir, ext) {
  fs.mkdirSync(extractDir, { recursive: true });
  if (ext === '.zip') {
    const psCmd = `Expand-Archive -Path "${archivePath}" -DestinationPath "${extractDir}" -Force`;
    execSync(`powershell.exe -Command "${psCmd}"`, { stdio: 'ignore' });
  } else {
    execSync(`tar xzf "${archivePath}" -C "${extractDir}"`, { stdio: 'ignore' });
  }
}

async function downloadBinary() {
  const version = getLatestVersion();
  const { ext } = getPlatform();
  const assetName = getAssetName();
  const archiveUrl = `https://github.com/${REPO}/releases/download/${version}/${assetName}`;
  const binPath = path.join(BIN_DIR, BIN_NAME + (process.platform === 'win32' ? '.exe' : ''));

  if (fs.existsSync(binPath)) {
    console.log(`Binary already exists at ${binPath}`);
    return;
  }

  const tmpDir = path.join(__dirname, '..', '.tmp');
  fs.mkdirSync(tmpDir, { recursive: true });
  const archivePath = path.join(tmpDir, assetName);

  console.log(`Downloading ${assetName}...`);
  await downloadFile(archiveUrl, archivePath);

  console.log(`Extracting ${assetName}...`);
  extractArchive(archivePath, tmpDir, ext);

  const extractedBin = path.join(tmpDir, BIN_NAME + (process.platform === 'win32' ? '.exe' : ''));
  if (!fs.existsSync(extractedBin)) {
    throw new Error(`Binary not found in archive at ${extractedBin}`);
  }

  fs.mkdirSync(BIN_DIR, { recursive: true });
  fs.renameSync(extractedBin, binPath);
  fs.chmodSync(binPath, 0o755);

  fs.rmSync(tmpDir, { recursive: true, force: true });

  console.log(`Binary installed at ${binPath}`);
}

downloadBinary().catch((err) => {
  console.error('Failed to install binary:', err.message);
  process.exit(0);
});
