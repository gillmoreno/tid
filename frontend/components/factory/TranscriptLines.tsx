interface TranscriptLinesProps {
  text: string;
}

export function TranscriptLines({ text }: TranscriptLinesProps) {
  const lines = text.split("\n").filter((line) => line.trim());

  return (
    <div className="space-y-2">
      {lines.map((line, index) => {
        const match = line.match(/^\[(\d{2}:\d{2}:\d{2})\]\s*(.*)$/);
        if (!match) {
          return (
            <p key={index} className="text-base leading-7 text-fog-light">
              {line}
            </p>
          );
        }

        return (
          <p key={index} className="text-base leading-7 text-fog-light">
            <span className="mr-2 inline-block min-w-[4.5rem] font-mono text-sm text-signal">
              {match[1]}
            </span>
            <span>{match[2]}</span>
          </p>
        );
      })}
    </div>
  );
}