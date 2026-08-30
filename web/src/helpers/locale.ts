import { registerLocale, toAlpha2 } from "@cospired/i18n-iso-languages";
import enLang from "@cospired/i18n-iso-languages/langs/en.json";

registerLocale(enLang);

export function parseBCP47(lang?: string): string {
  if (!lang) return "";
  let clean = lang.toLowerCase().trim();
  if (clean.includes("-")) {
    clean = clean.split("-")[0];
  }
  if (clean.includes("_")) {
    clean = clean.split("_")[0];
  }
  return clean;
}

export function get2LetterLangCode(lang?: string): string {
  if (!lang) return "";
  const parsed = parseBCP47(lang);
  const alpha2 = toAlpha2(parsed);
  if (alpha2) {
    return alpha2.toLowerCase();
  }
  return parsed.length === 2 ? parsed : parsed.slice(0, 2);
}