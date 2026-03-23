import "./ProviderProfiles.css";
import { useProviderProfiles } from "../../api/hooks/providerProfiles";
import { Button, Card, CardContent } from "@mui/material";
import { useDeleteProviderProfile } from "../../api/hooks/providerProfiles";
import toast from "react-hot-toast";

export default function ProviderProfiles() {
  const { data: providerProfiles, isLoading: isProviderProfilesLoading } =
    useProviderProfiles();
  return (
    <div>
      <h2>Provider Profiles</h2>
      <hr />
      {isProviderProfilesLoading ? (
        <div>Loading...</div>
      ) : (
        providerProfiles?.map((profile: any) => {
          return (
            <ProviderProfile key={profile.provider_id} profile={profile} />
          );
        })
      )}
    </div>
  );
}

function ProviderProfile({ profile }: { profile: any }) {
  const deleteProviderProfile = useDeleteProviderProfile();
  return (
    <Card
      variant="outlined"
      key={profile.provider_profile_id}
      className="mb-2 provider-profile-card"
    >
      <CardContent className="provider-profile-card-content">
        <h5>{profile.name}</h5>
        <div className="text-muted">{profile.manifest_url}</div>
        <Button
          variant="outlined"
          size="small"
          onClick={() => {
            deleteProviderProfile.mutate(profile.provider_profile_id, {
              onSuccess: () => {
                toast.success("Provider profile deleted");
              },
              onError: () => {
                toast.error("Failed to delete provider profile");
              },
            });
          }}
        >
          Delete
        </Button>
      </CardContent>
    </Card>
  );
}

function parseManifestURL(manifestURL: string) {
  const url = new URL(manifestURL);
  return url.origin;
}
