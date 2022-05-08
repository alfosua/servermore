import * as path from 'https://deno.land/std@0.138.0/path/mod.ts';
import { Application, Context } from "https://deno.land/x/abc@v1.3.3/mod.ts";
import { loadOptions } from "./args.ts";

const options = loadOptions();

const workerApp = new Application();

const userFunctions = await getUserFunctions(options.appDir);

workerApp
  .get("/internals", () => ({ functions: Object.keys(userFunctions) }))
  .get("/:funcName", handleFunctionCall)
  .start({ port: options.port });

type FunctionsMap = Record<string, (...args: unknown[]) => unknown>;

async function getUserFunctions(
  appDir: string,
): Promise<Readonly<FunctionsMap>> {
  const appEntryPointPath = path.join(appDir, "index.ts");
  const appEntryPointUrl = path.toFileUrl(appEntryPointPath);
  const userApp = await import(appEntryPointUrl.href);

  const functions: FunctionsMap = {};

  Object.entries(userApp)
    .filter(([_key, func]) => typeof func === "function")
    .forEach(([key, func]) => {
      if (typeof func === "function") {
        functions[key] = func as (...args: unknown[]) => unknown;
      }
    });

  return Object.freeze(functions);
}

async function handleFunctionCall(context: Context): Promise<unknown> {
  const funcName = context.params.funcName;

  if (funcName in userFunctions) {
    const func = userFunctions[funcName];

    const result = await func();

    return result;
  }

  return undefined;
}
