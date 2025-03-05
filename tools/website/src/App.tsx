import { useEffect, useState } from 'react';
import Search from './components/search';
import List, { VersionData } from './components/list/list';
import './App.css';

function App() {
    const [search, setSearch] = useState('');
    const [data, setData] = useState<VersionData>();
    useEffect(() => {
        getVersionData().then((data) => {
            setData(data);
        });
    }, []);
    return (
        <>
            <Search setSearch={setSearch} />
            <List data={data} search={search} />
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
