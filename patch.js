const fs = require('fs');

let pursFile = '/Users/0x1/Documents/htdocs/gopurs/gopurs-argonaut-core/src/Data/Argonaut/Core.purs';
let content = fs.readFileSync(pursFile, 'utf8');
content = content.replace(/foreign import _caseJson/g, 'foreign import caseJsonImpl');
content = content.replace(/_caseJson/g, 'caseJsonImpl');
fs.writeFileSync(pursFile, content);

let jsFile = '/Users/0x1/Documents/htdocs/gopurs/gopurs-argonaut-core/src/Data/Argonaut/Core.js';
let jsContent = fs.readFileSync(jsFile, 'utf8');
jsContent = jsContent.replace(/export const _caseJson =/g, 'export const caseJsonImpl =');
fs.writeFileSync(jsFile, jsContent);

let goFile = '/Users/0x1/Documents/htdocs/gopurs/gopurs-argonaut-core/src/Data/Argonaut/Core.go';
let goContent = fs.readFileSync(goFile, 'utf8');
goContent = goContent.replace(/func _CaseJson\(/g, 'func CaseJsonImpl(');
fs.writeFileSync(goFile, goContent);
