#!/usr/bin/env node

import puppeteer from "puppeteer";
import fs from "node:fs";
import os from "node:os";

const homedir = os.homedir();
const cache_file_date = `${homedir}/.cache/js_wotd_cache_ts.txt`;
const cache_file_word = `${homedir}/.cache/js_wotd_word.json`;
const today = new Date().toDateString();

const fetchWord = async () => {
  // Start a Puppeteer session with:
  // - a visible browser (`headless: false` - easier to debug because you'll see the browser in action)
  // - no default viewport (`defaultViewport: null` - website page will in full width and height)
  const browser = await puppeteer.launch({
    //headless: false,
    headless: true,
    defaultViewport: null,
  });

  const page = await browser.newPage();

  await page.goto("https://www.merriam-webster.com/word-of-the-day", {
    waitUntil: "domcontentloaded",
  });

  const word_data = await page.evaluate(() => {
    const word = document.querySelector("h2.word-header-txt").textContent;
    const word_partofspeech = document.querySelector(
      ".word-attributes span.main-attr",
    ).textContent;
    const word_syllables = document.querySelector(
      ".word-attributes span.word-syllables",
    ).textContent;
    const word_description = document.querySelectorAll(
      ".wod-definition-container p",
    )[0].textContent;
    const word_usage = document.querySelectorAll(
      ".wod-definition-container p",
    )[1].textContent;

    return {
      word,
      word_partofspeech,
      word_syllables,
      word_description,
      word_usage,
    };
  });
  // console.log(`### getting definition of ${word_data.word}`);
  // const word_description2 = fetch(
  //   `https://api.dictionaryapi.dev/api/v2/entries/en/${word_data[0]}`,
  // ).then((desc) => {
  //   console.log(desc);
  //   word_data["word_description2"] = desc;
  // });
  const getDefinition = async () => {
    const response = await fetch(
      ` https://api.dictionaryapi.dev/api/v2/entries/en/${word_data.word} `,
    );
    const json = await response.json();
    // console.log(json[0].meanings[0].definitions[0].definition);
    return json[0].meanings[0].definitions[0].definition;
  };

  getDefinition().then((word_definition) => {
    // console.log(word_definition);
    word_data.word_definition = word_definition;
    // console.log(word_data.word_definition);
    fs.writeFile(cache_file_word, JSON.stringify(word_data), function (err) {
      if (err) throw err;
      // console.log("### saved todays word to cache");
    });
  });

  await browser.close();
};

function readWord() {
  fs.readFile(cache_file_word, "utf8", (err, cache_word) => {
    if (err) {
      // console.error(err);
      fetchWord();
    }

    // console.log(cache_word);
    var cache_word_parsed = JSON.parse(cache_word);
    //     console.log(`${cache_word_parsed["word"]} (${cache_word_parsed["word_partofspeech"]} • ${cache_word_parsed["word_syllables"]})
    // ${cache_word_parsed["word_description"]}
    // ${cache_word_parsed["word_definition"]}
    // ${cache_word_parsed["word_usage"]}
    // `);
    console.log(`${cache_word_parsed["word"]} (${cache_word_parsed["word_partofspeech"]} • ${cache_word_parsed["word_syllables"]})

${cache_word_parsed["word_definition"]}
`);
  });
}

function main() {
  fs.readFile(cache_file_date, "utf8", (err, cache_date) => {
    if (err) {
      // console.error(err);
      fs.writeFile(cache_file_date, today, function (err) {
        if (err) throw err;
        // console.log("### saved todays date to cache");
      });
    }

    // console.log(cache_date);
    if (cache_date === new Date().toDateString()) {
      // console.log("### we should have this cached");
      readWord();
    } else {
      fetchWord(); //.then(readWord());
    }
  });
}

main();
