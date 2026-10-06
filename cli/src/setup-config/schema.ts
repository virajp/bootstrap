/** The setup config `.config/bootstrap.yaml`, equal to docs/blueprint/entities/setup-config/schema.yaml. */
import { Schema } from "effect";

const UniqueStrings = (item: Schema.String) => Schema.Array(item).check(Schema.isUnique());

/** One exact repository-relative file path; Validity checks it is literal and relative. */
const Paths = UniqueStrings(Schema.String);

const BranchModel = Schema.Literals(["direct", "pr"]);

export const Values = Schema.Struct({
  repo: Schema.String.check(Schema.isPattern(/^[^/\s]+(\/[^/\s]+)+$/)),
  commit_scopes: UniqueStrings(Schema.String.check(Schema.isPattern(/^[a-z][a-z0-9-]*$/))),
  merge_model: Schema.Struct({ develop: BranchModel, main: BranchModel }),
  tools: UniqueStrings(Schema.String),
});

export const SetupConfig = Schema.Struct({
  format: Schema.Int.check(Schema.isGreaterThanOrEqualTo(1)),
  version: Schema.String.check(
    Schema.isPattern(/^\d+\.\d+\.\d+(-[0-9A-Za-z.-]+)?(\+[0-9A-Za-z.-]+)?$/),
  ),
  values: Values,
  files: Paths,
  kept: Schema.optionalKey(Paths),
  deleted: Schema.optionalKey(Paths),
});

export type SetupConfig = typeof SetupConfig.Type;
