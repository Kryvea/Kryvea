import { mdiCog } from "@mdi/js";
import { useEffect, useState } from "react";
import { toast } from "react-toastify";
import { getData, putData } from "../api/api";
import Card from "../components/Composition/Card";
import Divider from "../components/Composition/Divider";
import Grid from "../components/Composition/Grid";
import PageHeader from "../components/Composition/PageHeader";
import Button from "../components/Form/Button";
import Buttons from "../components/Form/Buttons";
import Input from "../components/Form/Input";
import SelectWrapper from "../components/Form/SelectWrapper";
import type { Settings } from "../types/common.types";
import { languageMapping } from "../utils/constants";
import { getPageTitle } from "../utils/helpers";

const languageOptions = Object.entries(languageMapping).map(([code, label]) => ({
  value: code,
  label,
}));

export default function Settings() {
  const [settings, setSettings] = useState<Settings>({
    max_image_size: 0,
    default_category_language: "",
  });

  useEffect(() => {
    document.title = getPageTitle("Settings");
    getData("/api/admin/settings", setSettings);
  }, []);

  const selectedLanguageOption = languageOptions.find(opt => opt.value === settings.default_category_language);

  const handleSizeUpload = max_image_size => {
    setSettings(prev => ({ ...prev, max_image_size }));
  };

  const handleSubmit = () => {
    putData("/api/admin/settings", settings, () => {
      toast.success("Settings updated");
    });
  };

  return (
    <div>
      <PageHeader icon={mdiCog} title="Settings" />
      <Card>
        <Grid className="gap-4">
          <Input
            type="number"
            label="Max poc image file size (MB)"
            helperSubtitle="Required"
            placeholder="100 MB"
            id="image_upload_size"
            value={settings.max_image_size}
            onChange={handleSizeUpload}
          />
          <SelectWrapper
            label="Default language (new category)"
            id="language"
            options={languageOptions}
            value={selectedLanguageOption}
            onChange={option => setSettings(prev => ({ ...prev, default_category_language: option.value }))}
          />
        </Grid>
        <Divider />
        <Buttons>
          <Button text="Submit" onClick={handleSubmit} />
        </Buttons>
      </Card>
    </div>
  );
}
