const fs = require('fs');
const path = require('path');

const BIN_DIR = path.join(__dirname, '..', 'bin');

function cleanup() {
  try {
    if (fs.existsSync(BIN_DIR)) {
      fs.rmSync(BIN_DIR, { recursive: true, force: true });
      console.log('Removed binary directory');
    }
  } catch (err) {
    console.error('Cleanup failed:', err.message);
  }
}

cleanup();
