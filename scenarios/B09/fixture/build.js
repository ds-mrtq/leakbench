require('fs').mkdirSync('build', {recursive: true});
require('fs').writeFileSync('build/index.js', 'module.exports = 42;\n');
