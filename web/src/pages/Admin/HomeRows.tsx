import { useState } from "react";

import { useDefaultHomeRows } from "../../api/hooks/home";
import "./HomeRows.css";
import { Card, CardContent } from "@mui/material";

export default function HomeRows() {
  const { data, isLoading, error } = useDefaultHomeRows();
  const [isAddUserDialogOpen, setIsAddUserDialogOpen] = useState(false);

  if (isLoading) {
    return <div>Loading...</div>;
  }

  if (error) {
    return <div>Error: {error.message}</div>;
  }

  return (
    <div className="w-100">
      <h2>Default Home Catalogs</h2>
      <hr />
      <p>
        Change the default home catalogs for all users here. Users can customize
        their own home catalogs from their Account Settings page.
      </p>
      {data?.home_rows?.map((row: any) => (
        <div key={row.title} className="home-row">
          <div className="home-row-catalogs">
            {row.catalogs.map((catalog: any) => (
              <div key={catalog.catalog_id} className="home-row-catalog">
                <HomeRowCard homeRow={row} />
              </div>
            ))}
          </div>
        </div>
      ))}
    </div>
  );
}

function HomeRowCard({
  homeRow,
}: {
  homeRow: {
    title: string;
    catalogs: {
      catalog_id: string;
    }[];
  };
}) {
  return (
    <Card variant="outlined" className="mt-3">
      <CardContent>
        <div className="home-row-title">
          <p>Row Title: {homeRow.title}</p>
        </div>
        <div className="home-row-catalogs">
          {homeRow.catalogs.map((catalog) => (
            <div key={catalog.catalog_id} className="home-row-catalog">
              <p>{catalog.catalog_id}</p>
            </div>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}
