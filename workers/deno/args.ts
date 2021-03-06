import * as path from "https://deno.land/std@0.138.0/path/mod.ts";
import { parse } from "https://deno.land/std@0.138.0/flags/mod.ts";

export type Options = {
  hostUrl: URL,
  port: number;
  appDir: string;
};

export function loadOptions(): Options {
  const parsedArgs = parse(Deno.args, {
    default: {
      hostUrl: "http://localhost",
      port: 3000,
      appDir: "",
    },
  });

  if (!parsedArgs.appDir) {
    throw new Error("App directory is not defined");
  }

  return {
    hostUrl: new URL(parsedArgs.hostUrl),
    port: parsedArgs.port,
    appDir: path.resolve(parsedArgs.appDir),
  };
}
