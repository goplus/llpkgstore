import { VersionData } from '../../tools/parser/types';

interface VersionItemProps {
    ver: VersionData[string]['versions'][number];
    setVersion: (tag: string) => void;
}

const VersionItem: React.FC<VersionItemProps> = ({ ver, setVersion }) => {
    return (
        <div className="flex flex-row items-center gap-4 rounded-lg bg-gray-50 px-4 py-3">
            <span className="min-w-16 text-left text-lg font-bold">
                {ver.original}
            </span>
            <div className="flex flex-row gap-2">
                {ver.converted.map((ver, index) => (
                    <VersionTag key={index} tag={ver} onClick={setVersion} />
                ))}
            </div>
        </div>
    );
};

export default VersionItem;

interface VersionTagProps {
    tag: string;
    className?: string;
    onClick?: (tag: string) => void;
}

const VersionTag: React.FC<VersionTagProps> = ({ tag, className, onClick }) => {
    return (
        <button
            onClick={() => onClick && onClick(tag)}
            className={[
                'cursor-pointer rounded-full bg-gradient-to-r from-blue-50 to-purple-50 px-3 py-1 text-sm text-blue-600 transition-colors duration-300 hover:from-blue-100 hover:to-purple-100',
                className,
            ].join(' ')}
        >
            {tag}
        </button>
    );
};
