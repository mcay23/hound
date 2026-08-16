import React from "react";
import {
  Button,
  Checkbox,
  Dialog,
  DialogActions,
  FormControl,
  FormControlLabel,
  TextField,
} from "@mui/material";

export interface CollectionFormData {
  collection_title: string;
  description: string;
  is_public: boolean;
}

interface CollectionFormDialogProps {
  open: boolean;
  title: string;
  submitLabel: string;
  initialData: CollectionFormData;
  onClose: () => void;
  onSubmit: (data: CollectionFormData) => void;
}

export default function CollectionFormDialog({
  open,
  title,
  submitLabel,
  initialData,
  onClose,
  onSubmit,
}: CollectionFormDialogProps) {
  const [formData, setFormData] =
    React.useState<CollectionFormData>(initialData);

  React.useEffect(() => {
    setFormData(initialData);
  }, [initialData, open]);

  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const { name, type, value, checked } = event.target;
    setFormData((prev) => ({
      ...prev,
      [name]: type === "checkbox" ? checked : value,
    }));
  };

  const handleSubmit = () => {
    onSubmit(formData);
  };

  return (
    <Dialog
      open={open}
      onClose={onClose}
      aria-labelledby="collection-dialog-title"
      aria-describedby="collection-dialog-description"
    >
      <div className="reviews-create-dialog-header mb-3">{title}</div>
      <div className="reviews-create-dialog-content">
        <FormControl fullWidth>
          <TextField
            id="outlined-basic"
            label="Title"
            variant="outlined"
            name="collection_title"
            value={formData.collection_title}
            onChange={handleChange}
          />
          <TextField
            id="outlined-multiline-static"
            className="mt-3"
            label="Description"
            name="description"
            multiline
            rows={4}
            value={formData.description}
            onChange={handleChange}
          />
          <FormControlLabel
            className="mt-3"
            control={
              <Checkbox
                checked={formData.is_public}
                onChange={handleChange}
                name="is_public"
              />
            }
            label="Public Collection (Anyone can see)"
          />
        </FormControl>
      </div>
      <DialogActions>
        <Button onClick={onClose}>Cancel</Button>
        <Button onClick={handleSubmit}>{submitLabel}</Button>
      </DialogActions>
    </Dialog>
  );
}
