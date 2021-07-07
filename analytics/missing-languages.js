const fs = require("fs");
const es = require("event-stream");
const JSONStream = require("JSONStream");
const ProgressBar = require("progress");

let bar;
if (process.argv.length > 2) {
  const total = Number.parseInt(process.argv[2]);

  bar = new ProgressBar(" processing [:bar] :rate/s :percent :etas", {
    complete: "=",
    incomplete: " ",
    width: 20,
    total,
  });
}

function getStream() {
  var jsonData = "./out.json",
    stream = fs.createReadStream(jsonData, { encoding: "utf8" }),
    parser = JSONStream.parse("*");
  return stream.pipe(parser);
}

const unique = (value, index, self) => self.indexOf(value) === index;

const languages = [],
  missing = [];
let m = 0;

if (bar) {
  const timer = setInterval(() => {
    bar.tick();
    if (bar.complete) {
      clearInterval(timer);
    }
  }, 100);
}

getStream()
  .on("data", (data) => {
    const i = languages.findIndex((l) => l.language === data.speaker.language);
    if (i > -1) {
      languages[i].count++;
    } else {
      languages.push({
        language: data.speaker.language,
        count: 1,
      });
    }
    if (data.speaker.language === "") {
      if (!missing.includes(data.speaker.name)) {
        missing.push(data.speaker.name);
      }
      m++;
    }
    if (bar) {
      bar.tick();
    }
  })
  .on("close", () => {
    setTimeout(() => {
      fs.writeFileSync(
        "./stats.json",
        JSON.stringify(
          {
            sentencesWithoutLanguage: m,
            speakersWithoutLanguage: missing,
            languages,
          },
          null,
          2
        )
      );

      console.log("done.");
    }, 1000);
  });
