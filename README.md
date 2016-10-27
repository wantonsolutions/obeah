#Obeah
## Byzantine Fault Generation via SMT Guided Fuzz Testing

Obeah profiles the control flow of a single node in a distributed system. After sufficient control flow information has been collected, obeah queries an SMT to generate an unlikely control flow path. Using Symbolic execution obeah determines if any modification to a network payload can be made which would cause the node to execute the control path. This procedure is performed iteratively until a fail state is generated. Obeah outputs the set of message modifications which lead to the faulty state, and the control flow paths it forced the execution of.
