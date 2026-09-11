const fs = require('fs');
const data = JSON.parse(fs.readFileSync('output/Test.Main/corefn.json', 'utf8'));

function findF(node) {
    if (!node) return;
    if (node.type === 'App' && node.argument && node.argument.type === 'Var' && node.argument.value.identifier === 'cases') {
        console.log("Found APP with 'cases'");
        console.log("Argument 1 (cases):", JSON.stringify(node.argument));
        console.log("Function part:", JSON.stringify(node.abstraction, null, 2));
    }
    for (let key in node) {
        if (typeof node[key] === 'object') {
            findF(node[key]);
        }
    }
}
findF(data);
