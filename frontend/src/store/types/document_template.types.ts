export type DocumentTemplate = {
  id: string;
  name: string;
  code: string;
};

export type DocumentTemplateSheet = {
  id: string;
  templateId: string;
  code: string;
  name: string;
  required: boolean;
  headerRow: number;
  startRow: number;
};

export type DocumentTemplateColumn = {
  id: string;
  sheetId: string;
  key: string;
  name: string;
  dataType: string;
  required: boolean;
  isUnique: boolean;
  allowedValues?: string[];
};
