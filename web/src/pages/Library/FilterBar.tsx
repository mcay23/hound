import {
  Box,
  Button,
  FormControl,
  InputLabel,
  MenuItem,
  Select,
} from "@mui/material";

function FilterBar(props: any) {
  return (
    <Box textAlign="center" className="mx-auto mt-5 mb-4">
      <Button
        variant="contained"
        onClick={() => {
          props.showTitle
            ? props.setShowTitle(false)
            : props.setShowTitle(true);
        }}
      >
        My button
      </Button>
      <FormControl sx={{ minWidth: 120 }}>
        <InputLabel id="demo-simple-select-label">Age</InputLabel>
        <Select
          labelId="demo-simple-select-label"
          id="demo-simple-select"
          label="Age"
        >
          <MenuItem value={10}>Ten</MenuItem>
          <MenuItem value={20}>Twenty</MenuItem>
          <MenuItem value={30}>Thirty</MenuItem>
        </Select>
      </FormControl>
    </Box>
  );
}

export default FilterBar;
