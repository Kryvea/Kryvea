export type ObjectKey = string | number | symbol;

export type ObjectWithId = {
  id: string;
  [k: ObjectKey]: any;
};

export type Vulnerability = {
  id: string;
  updated_at: string;
  category: Category;
  detailed_title: string;
  status: string;
  cvssv2: {
    version: string;
    vector: string;
    score: number;
    severity: { label: string };
    description: string;
  };
  cvssv3: {
    version: string;
    vector: string;
    score: number;
    severity: { label: string };
    description: string;
  };
  cvssv31: {
    version: string;
    vector: string;
    score: number;
    severity: { label: string };
    description: string;
  };
  cvssv4: {
    version: string;
    vector: string;
    score: number;
    severity: { label: string };
    description: string;
  };
  customer: Customer;
  references: string[];
  generic_description: { enabled: boolean; text: string };
  generic_remediation: { enabled: boolean; text: string };
  description: string;
  remediation: string;
  target: { id: string; ipv4: string; ipv6: string; fqdn: string; name: string; customer: Customer };
  assessment: Assessment;
  user: { id: string; username: string };
};

export type User = {
  id: string;
  disabled_at: string;
  username: string;
  role: string;
  customers: Customer[];
  assessments: { id: string; name: string };
};

export type Assessment = {
  id: string;
  created_at: string;
  updated_at: string;
  name: string;
  start_date_time: string;
  end_date_time: string;
  kickoff_date_time: string;
  language: string;
  targets: Target[];
  status: string;
  type: { short: string; full: string };
  cvss_versions: { "2.0": boolean; "3.1": boolean; "4.0": boolean };
  environment: string;
  testing_type: string;
  osstmm_vector: string;
  vulnerability_count: number;
  customer: Customer;
  is_owned: boolean;
};

export type Target = {
  id: string;
  ipv4: string;
  ipv6: string;
  port: number;
  protocol: string;
  fqdn: string;
  tag: string;
  customer: Customer;
};

export type Customer = {
  id: string;
  name: string;
  language: string;
  logo_id: string;
  updated_at: Date;
  created_at: Date;
  templates: Template[];
};

export type Category = {
  id: string;
  updated_at: string;
  identifier: string;
  name: string;
  subcategory: string;
  source: "owasp_web" | "owasp_mobile" | "owasp_api" | "owasp_llm" | "att&ck" | "burp" | "cwe" | "nessus";
  generic_description: Record<string, string>;
  generic_remediation: Record<string, string>;
  languages_order: string[];
  references: string[];
};

export type Template = {
  id: string;
  name: string;
  filename: string;
  language: string;
  identifier: string;
  template_type: string;
  file_id: string;
  customer: Customer;
  created_at: Date;
};

export type Settings = {
  max_image_size: number;
  default_category_language: string;
};

export type ThemeMode = "light" | "dark" | "os";

export const exportTypes = [
  { value: "docx", label: "Word (.docx)" },
  { value: "xlsx", label: "Excel (.xlsx)" },
  { value: "zip-default", label: "Zip Archive (.zip)" },
];

export const uuidZero = "00000000-0000-0000-0000-000000000000";
