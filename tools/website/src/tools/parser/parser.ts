import { VersionData } from './types';

export const titleParser = (
  data?: VersionData,
  query?: string,
  page: number = 1,
  pageSize: number = 10,
): string[] => {
  if (!data) return [];
  const parsedData = Object.keys(data).filter((key) => {
    return query ? key.includes(query) : true;
  });

  const startIndex = (page - 1) * pageSize;
  const endIndex = startIndex + pageSize;

  return parsedData.slice(startIndex, endIndex);
};

export const versionParser = (
    data: VersionData[string],
    queryOrigin: string = '',
    queryMapped: string = '',
    page: number = 1,
    pageSize: number = 10
): VersionData[string]['versions'] => {
    const filteredVersions = data.versions.filter((ver) => {
        let flag = false;
        if (queryMapped.trim())
            ver.converted.forEach((con) => {
                if (con.includes(queryMapped.trim())) flag = true;
            });
        else flag = true;
        return (queryOrigin.trim() ? ver.original.includes(queryOrigin.trim()) : true) && flag;
    });

    const startIndex = (page - 1) * pageSize;
    const endIndex = startIndex + pageSize;

    return filteredVersions.slice(startIndex, endIndex);
};
