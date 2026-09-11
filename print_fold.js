const fs = require('fs');
const data = JSON.parse(fs.readFileSync('output/Test.Main/corefn.json', 'utf8'));

let foldNode = null;
for (let b of data.decls) {
  if (b.identifier === 'foldXXX') {
     foldNode = b;
  }
}
fs.writeFileSync('fold_ast.json', JSON.stringify(foldNode, null, 2));
