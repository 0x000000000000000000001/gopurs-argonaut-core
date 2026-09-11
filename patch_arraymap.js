const fs = require('fs');
let code = fs.readFileSync('../gopurs/src/Gopurs/CodeGen.purs', 'utf8');

code = code.replace(
  /if modNameStr == "Test.Main" then unsafePerformEffect \(Console.log \("APP ARRAYMAP in Test\.Main/g,
  'unsafePerformEffect (Console.log ("APP ARRAYMAP in " <> modNameStr'
);
code = code.replace(
  /if modNameStr == "Test.Main" then unsafePerformEffect \(Console.log \("UNCURRIED APP ARRAYMAP in Test\.Main/g,
  'unsafePerformEffect (Console.log ("UNCURRIED APP ARRAYMAP in " <> modNameStr'
);

fs.writeFileSync('../gopurs/src/Gopurs/CodeGen.purs', code);
