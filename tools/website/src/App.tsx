import { useEffect, useMemo, useState } from 'react';
import Header from './layout/header';
import List from './components/list/list';
import Search from './components/search';
import { titleParser } from './tools/parser/parser';
import { VersionData } from './tools/parser/types';
import './App.css';

function App() {
  const [search, setSearch] = useState('');
  const [data, setData] = useState<VersionData>();
  const searchResult = useMemo(
    () => titleParser(data, search, 1, 10),
    [data, search],
  );
  useEffect(() => {
    getVersionData().then((data) => {
      setData(data);
    });
  }, []);
  return (
    <>
      <Header />
      <Search
        query={search}
        setSearch={setSearch}
        dataNumber={data ? Object.keys(data).length : 0}
        resultNumber={searchResult.length}
      />
      <List data={data} titles={searchResult} />
    </>
  );
}

async function getVersionData(): Promise<VersionData> {
  const response = await fetch('./llpkgstore.json', {
    method: 'GET',
    headers: {
      'Cache-Control': 'no-cache',
    },
  });
  const data = await response.json();
  return data;
}

export default App;
