import { VersionData } from '../../tools/parser/types'

interface ItemProps {
    name: string
    data: VersionData[string]
    index: number
    setInfo: (data: string) => void
    setModalOpen: (open: boolean) => void
}

const Item: React.FC<ItemProps> = ({ name, data, setInfo, setModalOpen }) => {
    const remain = data.versions.length - 2
    return (
        <div className="flex min-h-32 flex-row items-center gap-4 overflow-clip rounded-xl border border-gray-300 bg-white p-4">
            <span className="w-32 text-2xl font-bold text-wrap text-gray-900">
                {name}
            </span>
            <div className="w-96 text-left">
                {data.versions
                    .filter((_, index) => {
                        return index < 2
                    })
                    .map((ver, index) => {
                        return (
                            <div
                                key={index}
                                className="flex flex-row items-center gap-4 text-nowrap overflow-ellipsis"
                            >
                                <span className="min-w-16 text-left text-lg font-bold overflow-ellipsis">
                                    {ver.original}
                                </span>
                                <span className="overflow-ellipsis">
                                    {ver.converted.join(' / ')}
                                </span>
                            </div>
                        )
                    })}
                {remain > 0 && (
                    <button
                        onClick={() => {
                            setInfo(name)
                            setModalOpen(true)
                        }}
                        className="cursor-pointer text-sky-500 hover:underline"
                    >
                        And {remain} more...
                    </button>
                )}
            </div>
        </div>
    )
}

export default Item
