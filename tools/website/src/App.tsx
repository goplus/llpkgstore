import { useEffect, useState } from 'react';
import Header from './layout/header'
import List, { VersionData } from './components/list/list'
import Search from './components/search'
import './App.css'

function App() {
    const [search, setSearch] = useState('')
    const [data, setData] = useState<VersionData>()
    useEffect(() => {
        getVersionData().then((data) => {
            setData(data)
        })
    }, [])
    return (
        <>
            <Header />
            <Search
                setSearch={setSearch}
                dataNumber={data ? Object.keys(data).length : 0}
            />
            <List data={data} search={search} />
        </>
    )
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
