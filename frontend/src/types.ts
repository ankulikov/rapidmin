export type AppConfig = {
  title: string;
  menu: MenuItem[];
  pages: Page[];
};

export type MenuItem = {
  title: string;
  page?: string;
  href?: string;
  children?: MenuItem[];
};

export type Page = {
  slug: string;
  title: string;
  widgets: Widget[];
};

export type Widget = {
  id: string;
  title: string;
  type: string;
  table?: TableSpec;
};

export type TableSpec = {
  columns: ColumnSpec[];
  filters?: FilterSpec[];
};

export type ColumnSpec = {
  id: string;
  title?: string;
  render?: ColumnRender;
};

export type ColumnRender = {
  type: string;
  text?: string;
  url?: string;
  external?: boolean;
};

export type FilterSpec = {
  id: string;
  title: string;
  type: string;
  target: string;
  operators?: string[];
  values?: ValueOption[];
};

export type ValueOption = {
  value: string;
  label: string;
};

export type DataResponse = {
  data: Record<string, unknown>[];
  total: number;
  next_cursor?: string;
  has_more?: boolean;
};
