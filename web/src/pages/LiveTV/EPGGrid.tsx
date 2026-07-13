import { useMemo, useRef } from "react";
import { Box, Paper, Typography, CircularProgress } from "@mui/material";
import { useLiveTVChannels, useChannelEPG } from "../../api/hooks/live_tv";

const CHANNEL_WIDTH = 220;
const ROW_HEIGHT = 70;
const PIXELS_PER_MINUTE = 4;
const SLOT_MINUTES = 30;

function getLocalizedText(items?: { text: string; lang?: string }[]): string {
  if (!items?.length) return "";

  return (
    items.find((x) => x.lang?.toLowerCase().startsWith("en"))?.text ??
    items[0].text
  );
}

function startOfHalfHour(date: Date) {
  const d = new Date(date);
  d.setMinutes(d.getMinutes() < 30 ? 0 : 30, 0, 0);
  return d;
}

function addMinutes(date: Date, minutes: number) {
  return new Date(date.getTime() + minutes * 60_000);
}

function minutesBetween(a: Date, b: Date) {
  return (b.getTime() - a.getTime()) / 60000;
}

function EPGChannelRow({
  iptvProfileID,
  channel,
  now,
  cutoff,
}: {
  iptvProfileID?: number;
  channel: any;
  now: Date;
  cutoff: Date;
}) {
  const { data: epg, isLoading } = useChannelEPG(
    iptvProfileID,
    channel.epg_channel_id,
  );

  const programmes =
    epg?.filter((p: any) => {
      const start = new Date(p.start_time);
      const stop = new Date(p.stop_time);

      return stop > now && start < cutoff;
    }) ?? [];

  return (
    <Box
      sx={{
        display: "flex",
        borderBottom: "1px solid",
        borderColor: "divider",
        height: ROW_HEIGHT,
      }}
    >
      {/* Channel column */}
      <Box
        sx={{
          width: CHANNEL_WIDTH,
          flexShrink: 0,
          borderRight: "1px solid",
          borderColor: "divider",
          display: "flex",
          alignItems: "center",
          px: 2,
          bgcolor: "background.paper",
          position: "sticky",
          left: 0,
          zIndex: 2,
        }}
      >
        <Typography noWrap>{channel.name}</Typography>
      </Box>

      {/* Timeline */}
      <Box
        sx={{
          position: "relative",
          width: minutesBetween(now, cutoff) * PIXELS_PER_MINUTE,
          flexShrink: 0,
        }}
      >
        {isLoading ? (
          <CircularProgress size={20} sx={{ mt: 2, ml: 2 }} />
        ) : (
          programmes.map((programme: any, index: number) => {
            const start = new Date(programme.start_time);
            const stop = new Date(programme.stop_time);

            const visibleStart = start < now ? now : start;

            const visibleStop = stop > cutoff ? cutoff : stop;

            const left = minutesBetween(now, visibleStart) * PIXELS_PER_MINUTE;

            const width =
              minutesBetween(visibleStart, visibleStop) * PIXELS_PER_MINUTE;

            return (
              <Paper
                key={index}
                sx={{
                  position: "absolute",
                  left,
                  top: 6,
                  width,
                  height: ROW_HEIGHT - 12,
                  overflow: "hidden",
                  px: 1,
                  py: 0.5,
                }}
              >
                <Typography variant="caption" fontWeight={600} noWrap>
                  {getLocalizedText(programme.titles)}
                </Typography>

                <Typography
                  variant="caption"
                  color="text.secondary"
                  display="block"
                  noWrap
                >
                  {new Date(programme.start_time).toLocaleTimeString([], {
                    hour: "2-digit",
                    minute: "2-digit",
                  })}
                  {" - "}
                  {new Date(programme.stop_time).toLocaleTimeString([], {
                    hour: "2-digit",
                    minute: "2-digit",
                  })}
                </Typography>
              </Paper>
            );
          })
        )}
      </Box>
    </Box>
  );
}

function EPGGrid({
  iptvProfileID,
  categoryID,
  hoursAhead = 12,
}: {
  iptvProfileID?: number;
  categoryID?: number;
  hoursAhead?: number;
}) {
  const { data: liveTVChannels } = useLiveTVChannels(iptvProfileID, categoryID);

  const scrollRef = useRef<HTMLDivElement>(null);

  const now = useMemo(() => new Date(), []);
  const cutoff = useMemo(
    () => new Date(now.getTime() + hoursAhead * 60 * 60 * 1000),
    [hoursAhead, now],
  );

  const timelineWidth = minutesBetween(now, cutoff) * PIXELS_PER_MINUTE;

  const slots = useMemo(() => {
    const result: Date[] = [];

    for (
      let t = startOfHalfHour(now);
      t <= cutoff;
      t = addMinutes(t, SLOT_MINUTES)
    ) {
      result.push(new Date(t));
    }

    return result;
  }, [now, cutoff]);

  return (
    <Paper sx={{ overflow: "hidden" }}>
      <Box
        ref={scrollRef}
        sx={{
          overflow: "auto",
        }}
      >
        {/* Header */}
        <Box
          sx={{
            display: "flex",
            position: "sticky",
            top: 0,
            bgcolor: "background.paper",
            zIndex: 5,
            borderBottom: "1px solid",
            borderColor: "divider",
          }}
        >
          <Box
            sx={{
              width: CHANNEL_WIDTH,
              flexShrink: 0,
              borderRight: "1px solid",
              borderColor: "divider",
              p: 2,
              position: "sticky",
              left: 0,
              zIndex: 6,
              bgcolor: "background.paper",
            }}
          >
            <Typography fontWeight={600}>Channel</Typography>
          </Box>

          <Box
            sx={{
              position: "relative",
              width: timelineWidth,
              height: 60,
              flexShrink: 0,
            }}
          >
            {slots.map((slot) => {
              const left = minutesBetween(now, slot) * PIXELS_PER_MINUTE;

              return (
                <Box
                  key={slot.toISOString()}
                  sx={{
                    position: "absolute",
                    left,
                    top: 0,
                    width: SLOT_MINUTES * PIXELS_PER_MINUTE,
                    height: "100%",
                    borderLeft: "1px solid",
                    borderColor: "divider",
                    px: 0.5,
                  }}
                >
                  <Typography variant="caption" fontWeight={600}>
                    {slot.toLocaleTimeString([], {
                      hour: "2-digit",
                      minute: "2-digit",
                    })}
                  </Typography>
                </Box>
              );
            })}
          </Box>
        </Box>

        {/* Rows */}
        {liveTVChannels
          ?.filter((c: any) => c.epg_channel_id)
          .map((channel: any) => (
            <EPGChannelRow
              key={channel.stream_id}
              iptvProfileID={iptvProfileID}
              channel={channel}
              now={now}
              cutoff={cutoff}
            />
          ))}
      </Box>
    </Paper>
  );
}

export default EPGGrid;
