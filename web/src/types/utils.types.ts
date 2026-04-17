/** Returns union of all keys types in T */
export type Keys<T> = keyof T;
/** Returns union of all values types in T */
export type Values<T> = T[Keys<T>];
/** Returns union of tuples containing [UNION OF KEYS TYPES, UNION OF VALUES TYPES] from each entry of T */
export type Entries<T> = [Keys<T>, Values<T>];
