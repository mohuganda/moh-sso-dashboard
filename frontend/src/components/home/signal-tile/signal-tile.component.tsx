import { Tile, Tag, Stack } from "@carbon/react";
import { getSeverityTagType, type Severity } from "../../../ui/severity";

type Props = {
  label: string;
  value: number | string;
  severity?: Severity;
  helperText?: string;
};

export function SignalTile({ label, value, severity, helperText }: Props) {
  return (
    <Tile className={`signal-tile ${severity ? `signal-tile--${severity}` : ""}`}>
      <Stack gap={2}>
        <span className="signal-tile__label">{label}</span>

        <div className="signal-tile__value">{value}</div>

        {(severity || helperText) && (
          <Stack orientation="horizontal" gap={2}>
            {severity && (
              <Tag size="sm" type={getSeverityTagType(severity)}>
                {severity}
              </Tag>
            )}

            {helperText && <span className="signal-tile__helper">{helperText}</span>}
          </Stack>
        )}
      </Stack>
    </Tile>
  );
}
