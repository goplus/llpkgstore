interface SearchProps {
    setSearch: (search: string) => void;
}

const Search: React.FC<SearchProps> = ({ setSearch }) => {
    return (
        <div className="mb-4 shadow-lg bg-white flex flex-row items-center px-8 justify-center">
            <span className="text-3xl font-bold absolute left-8">llpkgstore</span>
            <div className="px-4 py-2 border border-gray-300 rounded-3xl my-4 mx-8 flex items-center gap-2">
                <span className="text-xl font-bold">📦</span>
                <input
                    onChange={(e) => setSearch(e.target.value)}
                    className="focus-visible:outline-0"
                    placeholder="Search for a package"
                />
            </div>
            <a href="./versions.json" target="_blank" rel="noreferrer" className="text-blue-500">
                Raw JSON
            </a>
        </div>
    );
};

export default Search;
