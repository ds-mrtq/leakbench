const fs = require('fs');
const pkg = require('./package.json');
for (const dir of pkg.files) {
  if (!fs.existsSync(dir)) { console.error('publish failed: missing files directory ' + dir); process.exit(1); }
}
console.log('package contents accepted');
