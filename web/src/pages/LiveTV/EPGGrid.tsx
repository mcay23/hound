import { useMemo, useRef, useState, useEffect, useCallback } from "react";
import {
  useLiveTVChannels,
  useChannelEPGs,
  LiveTVChannel,
} from "../../api/hooks/live_tv";
import {
  useEpg,
  useProgram,
  Epg,
  Layout,
  ChannelBox,
  ChannelLogo,
  ProgramBox,
  ProgramContent,
  ProgramFlex,
  ProgramStack,
  ProgramTitle,
  ProgramText,
  ProgramItem,
} from "planby";

export interface EPGProgrammeLanguage {
  text: string;
  lang?: string;
}

export interface EPGProgramme {
  epg_channel_id: string;
  start_time: string;
  stop_time: string;
  titles: EPGProgrammeLanguage[];
  descriptions: EPGProgrammeLanguage[];
}

export interface SelectedChannel {
  channel: LiveTVChannel;
  epg: EPGProgramme[];
}

interface EPGGridProps {
  iptvProvider: any;
  categoryID?: number;
  selectedChannel: SelectedChannel | undefined;
  setSelectedChannel: (channel: SelectedChannel | undefined) => void;
  hoursAhead?: number;
}

// choose english or the first available language
export function pickText(items: EPGProgrammeLanguage[] | undefined): string {
  if (!items || items.length === 0) return "";
  const english = items.find(
    (i) => i.lang && i.lang.toLowerCase().startsWith("en"),
  );
  return (english ?? items[0]).text;
}

// wrap in responsive div, planby uses React 19 but we are
// backporting to React 18, so some functionality is broken
function EPGGrid(props: EPGGridProps) {
  const containerRef = useRef<HTMLDivElement>(null);
  const [width, setWidth] = useState<number | null>(null);

  useEffect(() => {
    if (!containerRef.current) return;
    setWidth(containerRef.current.clientWidth);
    const handleResize = () => {
      if (containerRef.current) {
        setWidth(containerRef.current.clientWidth);
      }
    };
    window.addEventListener("resize", handleResize);
    return () => {
      window.removeEventListener("resize", handleResize);
    };
  }, []);

  return (
    <div ref={containerRef} style={{ height: "600px", width: "100%" }}>
      {width !== null && (
        <EPGGridContent {...props} width={width} key={width} />
      )}
    </div>
  );
}

function EPGGridContent({
  iptvProvider: iptvProvider,
  categoryID,
  selectedChannel,
  setSelectedChannel,
  hoursAhead = 12,
  width,
}: EPGGridProps & { width: number }) {
  const { data: liveTVChannels } = useLiveTVChannels(iptvProvider, categoryID);
  const now = useMemo(() => new Date(), []);
  const before = useMemo(
    () => new Date(now.getTime() - 0.5 * 60 * 60 * 1000),
    [now],
  );
  const cutoff = useMemo(
    () => new Date(now.getTime() + hoursAhead * 60 * 60 * 1000),
    [hoursAhead, now],
  );
  const renderProgram = useCallback(
    (props: ProgramItem) => (
      <ProgramRenderer key={props.program.data.id} {...props} />
    ),
    [],
  );

  // Collect unique, non-empty epg_channel_ids from channels
  const epgChannelIDs = useMemo(() => {
    if (!liveTVChannels?.channels) return undefined;
    const ids = liveTVChannels.channels
      .map((ch: LiveTVChannel) => ch.epg_channel_id)
      .filter((id: string) => id && id.length > 0);
    return Array.from(new Set(ids)) as string[];
  }, [liveTVChannels]);

  const { data: rawEPGData } = useChannelEPGs(
    iptvProvider?.iptv_provider_id,
    iptvProvider?.iptv_provider_type,
    epgChannelIDs,
  );

  // map epg_channel_id to programmes
  const epgByChannelId = useMemo(() => {
    const map = new Map<string, EPGProgramme[]>();
    if (!rawEPGData) return map;
    for (const prog of rawEPGData as EPGProgramme[]) {
      const list = map.get(prog.epg_channel_id) || [];
      list.push(prog);
      map.set(prog.epg_channel_id, list);
    }
    return map;
  }, [rawEPGData]);

  // Helper to build a SelectedChannel from a LiveTVChannel
  const buildSelectedChannel = useCallback(
    (ch: LiveTVChannel): SelectedChannel => ({
      channel: ch,
      epg: epgByChannelId.get(ch.epg_channel_id) || [],
    }),
    [epgByChannelId],
  );

  const renderChannel = useCallback(
    (props: ChannelRendererProps) => (
      <ChannelRenderer
        key={props.channel.uuid + "|" + props.channel._liveTVChannel?.stream_id}
        {...props}
      />
    ),
    [],
  );

  useEffect(() => {
    if (
      !selectedChannel &&
      liveTVChannels?.channels &&
      liveTVChannels?.channels.length > 0
    ) {
      setSelectedChannel(buildSelectedChannel(liveTVChannels?.channels?.[0]));
    }
  }, [liveTVChannels, epgByChannelId]);

  // Convert live TV channels to planby channel format
  // Use epg_channel_id as uuid when available,
  // otherwise fall back to stream_id (epg data will not be available)
  const channels = useMemo(() => {
    if (!liveTVChannels?.channels) return [];
    return liveTVChannels.channels.map((ch: LiveTVChannel) => ({
      logo: ch.thumbnail_url,
      uuid: ch.epg_channel_id || String(ch.stream_id),
      name: ch.name,
      _liveTVChannel: ch,
    }));
  }, [liveTVChannels]);

  // Convert EPG programmes to planby program format.
  // dates from the api already contain timezone offsets (e.g. "+02:00"),
  // so new Date() correctly parses them to the browser's local time
  const epg = useMemo(() => {
    if (!rawEPGData) return [];
    return (rawEPGData as EPGProgramme[]).map(
      (prog: EPGProgramme, index: number) => {
        const since = new Date(prog.start_time);
        const till = new Date(prog.stop_time);
        return {
          channelUuid: prog.epg_channel_id,
          id: `${prog.epg_channel_id}-${prog.start_time}-${index}`,
          title: pickText(prog.titles),
          description: pickText(prog.descriptions),
          since: since.toISOString(),
          till: till.toISOString(),
          image: "",
        };
      },
    );
  }, [rawEPGData]);

  const { getEpgProps, getLayoutProps } = useEpg({
    epg,
    channels,
    startDate: before,
    endDate: cutoff.toISOString(),
    dayWidth: 3600,
    width,
    height: 600,
    sidebarWidth: 300,
  });

  return (
    <Epg {...getEpgProps()}>
      <Layout
        {...getLayoutProps()}
        renderChannel={(props) =>
          renderChannel({
            ...props,
            selectedChannel,
            setSelectedChannel: (ch: LiveTVChannel) =>
              setSelectedChannel(buildSelectedChannel(ch)),
          })
        }
        renderProgram={renderProgram}
      />
    </Epg>
  );
}

interface ChannelRendererProps {
  channel: any;
  selectedChannel: SelectedChannel | undefined;
  setSelectedChannel: (channel: LiveTVChannel) => void;
}

function ChannelRenderer({
  channel,
  selectedChannel,
  setSelectedChannel,
}: ChannelRendererProps) {
  const isSelected =
    selectedChannel?.channel.stream_url === channel._liveTVChannel?.stream_url;

  return (
    <ChannelBox {...channel.position}>
      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: 12,
          padding: "0 12px",
          width: "100%",
        }}
      >
        {channel.logo && (
          <ChannelLogo
            src={channel.logo}
            alt={channel.name}
            style={{
              maxWidth: 40,
              maxHeight: 40,
              objectFit: "contain",
            }}
          />
        )}
        <span
          style={{
            color: isSelected ? "yellow" : "#fff",
            fontSize: 14,
            fontWeight: 500,
            overflow: "hidden",
            textOverflow: "ellipsis",
            whiteSpace: "nowrap",
            width: "100%",
            cursor: "pointer",
          }}
          onClick={() => {
            const liveCh: LiveTVChannel = channel._liveTVChannel;
            setSelectedChannel(liveCh);
          }}
        >
          {channel.name}
        </span>
      </div>
    </ChannelBox>
  );
}

interface ProgramRendererProps extends ProgramItem {}
function ProgramRenderer(props: ProgramRendererProps) {
  const { program } = props;
  const { styles, formatTime, isLive } = useProgram(props);
  const { title, since, till } = program.data;

  return (
    <ProgramBox width={styles.width} style={styles.position}>
      <ProgramContent width={styles.width} isLive={isLive}>
        <ProgramFlex>
          <ProgramStack>
            <ProgramTitle>{title}</ProgramTitle>
            <ProgramText>
              {formatTime(since)} - {formatTime(till)}
            </ProgramText>
          </ProgramStack>
        </ProgramFlex>
      </ProgramContent>
    </ProgramBox>
  );
}

export default EPGGrid;
