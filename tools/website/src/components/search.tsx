interface SearchProps {
    dataNumber: number;
    resultNumber: number;
    query: string;
    setSearch: (search: string) => void;
}

const Search: React.FC<SearchProps> = ({
    dataNumber,
    resultNumber,
    query,
    setSearch,
}) => {
    return (
        <div className="flex w-full flex-col items-center p-6">
            <div className="mx-2 my-4 flex w-full items-center gap-2 rounded-xl border border-gray-300 bg-white px-4 py-2 sm:mx-8 md:w-fit">
                <span className="text-xl font-bold select-none">📦</span>
                <input
                    onChange={(e) => setSearch(e.target.value)}
                    className="w-full focus-visible:outline-0 md:w-96"
                    placeholder="Search for a package"
                />
            </div>
            <p>
                {query.trim() ? (
                    <>
                        {resultNumber ? (
                            <span className="font-bold text-blue-700">
                                {resultNumber}{' '}
                            </span>
                        ) : (
                            'No '
                        )}
                        results found from
                        <span className="font-bold text-blue-700">
                            {' '}
                            {dataNumber}{' '}
                        </span>
                        packages.
                    </>
                ) : (
                    <>
                        {dataNumber ? (
                            <span className="font-bold text-blue-700">
                                {dataNumber}{' '}
                            </span>
                        ) : (
                            'No '
                        )}
                        packages found.
                    </>
                )}
            </p>
        </div>
    );
};

export default Search;
