import { VersionData } from './list';

interface ItemProps {
    name: string;
    data: VersionData[string];
    index: number;
    setInfo: (data: string) => void;
    setModalOpen: (open: boolean) => void;
}

const Item: React.FC<ItemProps> = ({ name, data, setInfo, setModalOpen }) => {
    const remain = data.versions.length - 2;
    return (
        <div className="bg-white flex flex-row gap-4 items-center border rounded-xl border-gray-300 p-4 min-h-32">
            <span className="font-bold text-2xl w-32 text-wrap text-gray-900">{name}</span>
            <div className="w-96 text-left">
                {data.versions
                    .filter((_, index) => {
                        return index < 2;
                    })
                    .map((ver, index) => {
                        return (
                            <div key={index} className="flex flex-row gap-4 items-center">
                                <span className="font-bold text-lg min-w-16 text-left">{ver.original}</span>
                                <span>{ver.converted.join(' / ')}</span>
                            </div>
                        );
                    })}
                {remain > 0 && (
                    <button
                        onClick={() => {
                            setInfo(name);
                            setModalOpen(true);
                        }}
                        className="text-sky-500 cursor-pointer hover:underline">
                        And {remain} more...
                    </button>
                )}
            </div>
        </div>
    );
};

export default Item;
